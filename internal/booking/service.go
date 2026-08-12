package booking

import "context"

// Service provides booking operations using a [BookingStore].
type Service struct {
	store BookingStore
}

// NewService returns a new booking service backed by the given store.
func NewService(store BookingStore) *Service {
	return &Service{store: store}
}

// Book creates a booking using the configured [BookingStore].
func (s *Service) Book(b Booking) (Booking, error) {
	return s.store.Book(b)
}

// ListBookings returns all bookings for the given movie.
func (s *Service) ListBookings(movieID string) []Booking {
	return s.store.ListBookings(movieID)
}

// ConfirmSeat confirms a booking session for the given user.
func (s *Service) ConfirmSeat(
	ctx context.Context,
	sessionID string,
	userID string,
) (Booking, error) {
	return s.store.Confirm(ctx, sessionID, userID)
}

// ReleaseSeat releases a booking session for the given user.
func (s *Service) ReleaseSeat(
	ctx context.Context,
	sessionID string,
	userID string,
) error {
	return s.store.Release(ctx, sessionID, userID)
}
