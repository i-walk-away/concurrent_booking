package booking

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InMemoryStore stores bookings in memory.
//
// It is intended primarily for unit tests and local development where a
// persistent Redis instance is not required.
//
// InMemoryStore is safe for concurrent use and follows the same booking
// semantics as RedisStore, including temporary holds and session ownership.
type InMemoryStore struct {
	mu       sync.RWMutex
	bookings map[string]Booking
}

// NewInMemoryStore returns an empty in-memory booking store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		bookings: make(map[string]Booking),
	}
}

// Book creates a temporary booking for a seat.
//
// The booking expires after defaultHoldTTL. Expired holds are treated as
// available and may be replaced by a new booking.
func (s *InMemoryStore) Book(b Booking) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for sessionID, booking := range s.bookings {
		if s.isExpired(booking, now) {
			delete(s.bookings, sessionID)
			continue
		}

		if booking.MovieID == b.MovieID && booking.SeatID == b.SeatID {
			return Booking{}, ErrSeatAlreadyBooked
		}
	}

	b.ID = uuid.NewString()
	b.Status = statusHeld
	b.ExpiresAt = now.Add(defaultHoldTTL)

	s.bookings[b.ID] = b

	return b, nil
}

// ListBookings returns all active bookings for the given movie.
//
// Expired bookings are not returned.
func (s *InMemoryStore) ListBookings(movieID string) []Booking {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	bookings := make([]Booking, 0)

	for sessionID, booking := range s.bookings {
		if s.isExpired(booking, now) {
			delete(s.bookings, sessionID)
			continue
		}

		if booking.MovieID == movieID {
			bookings = append(bookings, booking)
		}
	}

	return bookings
}

// Confirm converts a temporary booking into a permanent booking.
func (s *InMemoryStore) Confirm(
	_ context.Context,
	sessionID string,
	userID string,
) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	booking, ok := s.bookings[sessionID]
	if !ok {
		return Booking{}, ErrSessionNotFound
	}

	if booking.UserID != userID {
		return Booking{}, ErrSessionNotOwned
	}

	if s.isExpired(booking, time.Now()) {
		delete(s.bookings, sessionID)
		return Booking{}, ErrSessionNotFound
	}

	booking.Status = statusConfirmed
	booking.ExpiresAt = time.Time{}

	s.bookings[sessionID] = booking

	return booking, nil
}

// Release removes a temporary booking.
func (s *InMemoryStore) Release(
	_ context.Context,
	sessionID string,
	userID string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	booking, ok := s.bookings[sessionID]
	if !ok {
		return ErrSessionNotFound
	}

	if booking.UserID != userID {
		return ErrSessionNotOwned
	}

	if s.isExpired(booking, time.Now()) {
		delete(s.bookings, sessionID)
		return ErrSessionNotFound
	}

	delete(s.bookings, sessionID)

	return nil
}

func (s *InMemoryStore) isExpired(booking Booking, now time.Time) bool {
	return booking.Status == statusHeld &&
		!booking.ExpiresAt.IsZero() &&
		!now.Before(booking.ExpiresAt)
}
