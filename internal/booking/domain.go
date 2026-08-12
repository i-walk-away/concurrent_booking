package booking

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrSeatAlreadyBooked indicates that the requested seat is already booked.
	ErrSeatAlreadyBooked = errors.New("seat is already taken")
)

// Booking represents a confirmed seat reservation.
type Booking struct {
	ID        string
	MovieID   string
	SeatID    string
	UserID    string
	Status    string
	ExpiresAt time.Time
}

// BookingStore defines the storage operations for seat bookings.
type BookingStore interface {
	// Book adds a booking to the store.
	//
	// It returns ErrSeatAlreadyBooked if the seat is already booked.
	Book(b Booking) (Booking, error)

	// ListBookings returns all bookings for the given movie.
	ListBookings(movieID string) []Booking

	// Confirm confirms a booking for the given session and user.
	Confirm(ctx context.Context, sessionID string, userID string) (Booking, error)

	// Release releases a booking for the given session and user.
	Release(ctx context.Context, sessionID string, userID string) error
}
