package booking

type MemoryStore struct {
	bookings map[string]Booking
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		bookings: map[string]Booking{},
	}
}

func (store *MemoryStore) Book(booking Booking) error {
	// check if the seat is already taken
	if _, exists := store.bookings[booking.SeatID]; exists == true {
		return ErrSeatAlreadyTaken
	}
	store.bookings[booking.SeatID] = booking
	return nil
}

func (store *MemoryStore) ListBookings(movieID string) []Booking {
	var result []Booking
	for _, booking := range store.bookings {
		if booking.MovieID == movieID {
			result = append(result, booking)
		}
	}
	return result
}
