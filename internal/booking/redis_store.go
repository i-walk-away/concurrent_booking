package booking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const defaultHoldTTL = 2 * time.Minute

// RedisStore implements session-based seat booking backed by Redis.
//
// Key design:
//
//	seat:{movieID}:{seatID}   → booking JSON (TTL = held, no TTL = confirmed)
//	session:{sessionID}       → seat key     (reverse lookup)
type RedisStore struct {
	rdb *redis.Client
}

// NewRedisStore returns a Redis-backed booking store using the given client.
func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

// sessionKey builds the reverse-lookup key for a session.
func sessionKey(id string) string {
	return fmt.Sprintf("session:%s", id)
}

// seatKey builds the Redis key for a movie seat.
func seatKey(movieID, seatID string) string {
	return fmt.Sprintf("seat:%s:%s", movieID, seatID)
}

// Book creates a temporary seat reservation.
//
// Redis SET NX provides the atomic check-and-set operation required to ensure
// that concurrent attempts to book the same seat result in at most one winner.
func (s *RedisStore) Book(b Booking) (Booking, error) {
	now := time.Now()

	b.ID = uuid.NewString()
	b.Status = statusHeld
	b.ExpiresAt = now.Add(defaultHoldTTL)

	key := seatKey(b.MovieID, b.SeatID)

	data, err := json.Marshal(b)
	if err != nil {
		return Booking{}, fmt.Errorf("marshal booking: %w", err)
	}

	ctx := context.Background()

	ok, err := s.rdb.SetNX(ctx, key, data, defaultHoldTTL).Result()
	if err != nil {
		return Booking{}, fmt.Errorf("hold seat: %w", err)
	}

	if !ok {
		return Booking{}, ErrSeatAlreadyBooked
	}

	if err := s.rdb.Set(
		ctx,
		sessionKey(b.ID),
		key,
		defaultHoldTTL,
	).Err(); err != nil {
		// Do not leave a seat hold behind if its reverse lookup could not
		// be created.
		_ = s.rdb.Del(ctx, key).Err()

		return Booking{}, fmt.Errorf("create session: %w", err)
	}

	return b, nil
}

// ListBookings returns all active bookings for the given movie.
//
// Redis SCAN is used instead of KEYS so that listing bookings does not block
// Redis while scanning a large keyspace.
func (s *RedisStore) ListBookings(movieID string) []Booking {
	ctx := context.Background()
	pattern := fmt.Sprintf("seat:%s:*", movieID)

	bookings := make([]Booking, 0)

	iter := s.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		value, err := s.rdb.Get(ctx, iter.Val()).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}

			continue
		}

		booking, err := parseBooking(value)
		if err != nil {
			continue
		}

		bookings = append(bookings, booking)
	}

	return bookings
}

// Confirm converts a held session into a permanent booking.
//
// Removing the TTL makes both the seat and session persistent.
func (s *RedisStore) Confirm(
	ctx context.Context,
	sessionID string,
	userID string,
) (Booking, error) {
	session, key, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return Booking{}, err
	}

	session.Status = statusConfirmed
	session.ExpiresAt = time.Time{}

	data, err := json.Marshal(session)
	if err != nil {
		return Booking{}, fmt.Errorf("marshal confirmed booking: %w", err)
	}

	if err := s.rdb.Set(ctx, key, data, 0).Err(); err != nil {
		return Booking{}, fmt.Errorf("confirm booking: %w", err)
	}

	if err := s.rdb.Persist(ctx, sessionKey(sessionID)).Err(); err != nil {
		return Booking{}, fmt.Errorf("persist session: %w", err)
	}

	return session, nil
}

// Release removes a booking session.
func (s *RedisStore) Release(
	ctx context.Context,
	sessionID string,
	userID string,
) error {
	_, key, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return err
	}

	if err := s.rdb.Del(
		ctx,
		key,
		sessionKey(sessionID),
	).Err(); err != nil {
		return fmt.Errorf("release session: %w", err)
	}

	return nil
}

func (s *RedisStore) getSession(
	ctx context.Context,
	sessionID string,
	userID string,
) (Booking, string, error) {
	key, err := s.rdb.Get(ctx, sessionKey(sessionID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Booking{}, "", ErrSessionNotFound
		}

		return Booking{}, "", fmt.Errorf("get session: %w", err)
	}

	value, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Booking{}, "", ErrSessionNotFound
		}

		return Booking{}, "", fmt.Errorf("get booking: %w", err)
	}

	session, err := parseBooking(value)
	if err != nil {
		return Booking{}, "", fmt.Errorf("parse booking: %w", err)
	}

	if session.UserID != userID {
		return Booking{}, "", ErrSessionNotOwned
	}

	return session, key, nil
}

func parseBooking(value string) (Booking, error) {
	var booking Booking

	if err := json.Unmarshal([]byte(value), &booking); err != nil {
		return Booking{}, err
	}

	return booking, nil
}
