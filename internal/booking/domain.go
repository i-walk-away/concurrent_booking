package booking

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrSeatAlreadyBooked indicates that the requested seat is already booked.
	ErrSeatAlreadyBooked = errors.New("seat is already taken")

	// ErrSessionNotFound indicates that the requested booking session does not exist
	// or has already expired.
	ErrSessionNotFound = errors.New("booking session not found")

	// ErrSessionNotOwned indicates that the booking session belongs to another user.
	ErrSessionNotOwned = errors.New("booking session does not belong to user")
)

const (
	statusHeld      = "held"
	statusConfirmed = "confirmed"
)

// Booking represents a seat reservation.
type Booking struct {
	ID        string
	MovieID   string
	SeatID    string
	UserID    string
	Status    string
	ExpiresAt time.Time
}

// BookingStore defines storage operations for seat bookings.
type BookingStore interface {
	// Book creates a temporary booking for a seat.
	//
	// It returns ErrSeatAlreadyBooked if the seat is already held or confirmed.
	Book(b Booking) (Booking, error)

	// ListBookings returns all active bookings for the given movie.
	ListBookings(movieID string) []Booking

	// Confirm converts a temporary booking into a permanent booking.
	//
	// ErrSessionNotFound is returned when the session does not exist or has expired.
	// ErrSessionNotOwned is returned when the session belongs to another user.
	Confirm(ctx context.Context, sessionID string, userID string) (Booking, error)

	// Release removes a temporary booking.
	//
	// ErrSessionNotFound is returned when the session does not exist or has expired.
	// ErrSessionNotOwned is returned when the session belongs to another user.
	Release(ctx context.Context, sessionID string, userID string) error
}
