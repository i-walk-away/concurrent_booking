package booking

import (
	"context"
	"errors"
	"sync"
	"time"
)

var errSessionNotFound = errors.New("booking session not found")

// MemoryStore stores bookings in memory.
//
// MemoryStore is safe for concurrent use.
type MemoryStore struct {
	mu       sync.RWMutex
	bookings map[string]Booking
}

// NewMemoryStore returns an empty in-memory booking store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		bookings: make(map[string]Booking),
	}
}

// Book adds a booking to the store.
//
// It returns ErrSeatAlreadyBooked if the seat is already booked.
func (s *MemoryStore) Book(b Booking) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, booking := range s.bookings {
		if booking.MovieID == b.MovieID && booking.SeatID == b.SeatID {
			return Booking{}, ErrSeatAlreadyBooked
		}
	}

	s.bookings[b.ID] = b

	return b, nil
}

// ListBookings returns all bookings for the given movie.
func (s *MemoryStore) ListBookings(movieID string) []Booking {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bookings := make([]Booking, 0)

	for _, booking := range s.bookings {
		if booking.MovieID == movieID {
			bookings = append(bookings, booking)
		}
	}

	return bookings
}

// Confirm confirms a booking session for the given user.
func (s *MemoryStore) Confirm(
	_ context.Context,
	sessionID string,
	userID string,
) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	booking, ok := s.bookings[sessionID]
	if !ok {
		return Booking{}, errSessionNotFound
	}

	if booking.UserID != userID {
		return Booking{}, errSessionNotFound
	}

	booking.Status = statusConfirmed
	booking.ExpiresAt = time.Time{}

	s.bookings[sessionID] = booking

	return booking, nil
}

// Release releases a booking session for the given user.
func (s *MemoryStore) Release(
	_ context.Context,
	sessionID string,
	userID string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	booking, ok := s.bookings[sessionID]
	if !ok {
		return errSessionNotFound
	}

	if booking.UserID != userID {
		return errSessionNotFound
	}

	delete(s.bookings, sessionID)

	return nil
}
