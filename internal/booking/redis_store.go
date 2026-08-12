package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const defaultHoldTTL = 2 * time.Minute

// RedisStore implements session-based seat booking backed by Redis.
//
// Seat keys store booking sessions and expire automatically while a seat is
// held. Confirmed bookings have no expiration.
//
// The store uses the following Redis keys:
//
//	seat:{movieID}:{seatID} → booking session
//	session:{sessionID}     → seat key for reverse lookup
type RedisStore struct {
	rdb *redis.Client
}

// NewRedisStore returns a Redis-backed booking store using the given client.
func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

// sessionKey returns the Redis key used for reverse lookup by session ID.
func sessionKey(id string) string {
	return fmt.Sprintf("session:%s", id)
}

// Book creates a temporary seat reservation.
//
// The reservation expires after the default hold duration. It returns
// ErrSeatAlreadyBooked if the seat is already reserved.
func (s *RedisStore) Book(b Booking) (Booking, error) {
	session, err := s.hold(b)
	if err != nil {
		return Booking{}, err
	}

	log.Printf("Session booked %v", session)

	return session, nil
}

// ListBookings returns all booking sessions for the given movie.
func (s *RedisStore) ListBookings(movieID string) []Booking {
	pattern := fmt.Sprintf("seat:%s:*", movieID)
	var sessions []Booking

	ctx := context.Background()

	iter := s.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		val, err := s.rdb.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}

		session, err := parseSession(val)
		if err != nil {
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions
}

func (s *RedisStore) hold(b Booking) (Booking, error) {
	id := uuid.New().String()
	now := time.Now()
	ctx := context.Background()
	key := fmt.Sprintf("seat:%s:%s", b.MovieID, b.SeatID)

	b.ID = id

	val, _ := json.Marshal(b)

	res := s.rdb.SetArgs(ctx, key, val, redis.SetArgs{
		Mode: "NX",
		TTL:  defaultHoldTTL,
	})

	if res.Val() != "OK" {
		return Booking{}, ErrSeatAlreadyBooked
	}

	s.rdb.Set(ctx, sessionKey(id), key, defaultHoldTTL)

	return Booking{
		ID:        id,
		MovieID:   b.MovieID,
		SeatID:    b.SeatID,
		UserID:    b.UserID,
		Status:    "held",
		ExpiresAt: now.Add(defaultHoldTTL),
	}, nil
}

func parseSession(val string) (Booking, error) {
	var data Booking
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return Booking{}, err
	}

	return Booking{
		ID:      data.ID,
		MovieID: data.MovieID,
		SeatID:  data.SeatID,
		UserID:  data.UserID,
		Status:  data.Status,
	}, nil
}

// Confirm confirms a held booking session for the given user.
//
// A confirmed booking is made permanent by removing the expiration time from
// both the seat key and the session key.
func (s *RedisStore) Confirm(
	ctx context.Context,
	sessionID string,
	userID string,
) (Booking, error) {
	session, sk, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return Booking{}, err
	}

	s.rdb.Persist(ctx, sk)
	s.rdb.Persist(ctx, sessionKey(sessionID))

	session.Status = "confirmed"

	data := Booking{
		ID:      session.ID,
		MovieID: session.MovieID,
		SeatID:  session.SeatID,
		UserID:  session.UserID,
		Status:  "confirmed",
	}

	val, _ := json.Marshal(data)
	s.rdb.Set(ctx, sk, val, 0)

	return session, nil
}

func (s *RedisStore) getSession(
	ctx context.Context,
	sessionID string,
	userID string,
) (Booking, string, error) {
	sk, err := s.rdb.Get(ctx, sessionKey(sessionID)).Result()
	if err != nil {
		return Booking{}, "", err
	}

	val, err := s.rdb.Get(ctx, sk).Result()
	if err != nil {
		return Booking{}, "", err
	}

	session, err := parseSession(val)
	if err != nil {
		return Booking{}, "", err
	}

	return session, sk, nil
}

// Release releases a held booking session for the given user.
func (s *RedisStore) Release(ctx context.Context, sessionID string, userID string) error {
	_, sk, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return err
	}

	s.rdb.Del(ctx, sk, sessionKey(sessionID))

	return nil
}
