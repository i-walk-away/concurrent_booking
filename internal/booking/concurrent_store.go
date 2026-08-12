package booking

import "sync"

// ConcurrentStore stores bookings in memory and is safe for concurrent use.
type ConcurrentStore struct {
	bookings map[string]Booking
	sync.RWMutex
}

// NewConcurrentStore returns an empty concurrent booking store.
func NewConcurrentStore() *ConcurrentStore {
	return &ConcurrentStore{
		bookings: map[string]Booking{},
	}
}

// Book adds a booking to the store.
//
// It returns ErrSeatAlreadyBooked if the seat is already booked.
func (s *ConcurrentStore) Book(b Booking) error {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.bookings[b.SeatID]; exists {
		return ErrSeatAlreadyBooked
	}

	s.bookings[b.SeatID] = b

	return nil
}

// ListBookings returns all bookings for the given movie.
func (s *ConcurrentStore) ListBookings(movieID string) []Booking {
	s.RLock()
	defer s.RUnlock()

	var result []Booking

	for _, b := range s.bookings {
		if b.MovieID == movieID {
			result = append(result, b)
		}
	}

	return result
}
