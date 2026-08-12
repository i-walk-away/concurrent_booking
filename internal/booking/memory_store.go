package booking

import "sync"

// MemoryStore is an in-memory implementation of the BookingStore interface.
// It stores all bookings in a map where the key is the SeatID and the value
// is the complete Booking record.
//
// Important characteristics:
//   - Not thread-safe, concurrent access requires external synchronization
//   - Data is lost when the process terminates (no persistence)
//   - Fast read/write operations (O(1) for single booking operations)
//   - Memory usage grows linearly with the number of bookings
type MemoryStore struct {
	bookings map[string]Booking
	sync.RWMutex
}

// NewMemoryStore creates and initializes a new MemoryStore instance.
// The store is created with an empty bookings map.
//
// Returns:
//   - *MemoryStore: pointer to the newly created MemoryStore instance
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		bookings: map[string]Booking{},
	}
}

// Book attempts to reserve a seat in the memory store.
// It implements the BookingStore interface method.
//
// The operation is performed in three steps:
//  1. Check if the SeatID already exists in the bookings map
//  2. If exists, return ErrSeatAlreadyTaken immediately
//  3. Otherwise, store the booking with SeatID as the key
//
// Parameters:
//   - booking: the Booking to be stored
//
// Returns:
//   - error: nil on successful booking, ErrSeatAlreadyTaken if the seat
//     is already reserved.
func (store *MemoryStore) Book(booking Booking) error {
	store.Lock()
	defer store.Unlock()

	// check if the seat is already taken
	if _, exists := store.bookings[booking.SeatID]; exists == true {
		return ErrSeatAlreadyTaken
	}
	store.bookings[booking.SeatID] = booking

	return nil
}

// ListBookings retrieves all bookings for a specific movie from the memory store.
// It implements the BookingStore interface method.
//
// The method iterates through all bookings in the store and filters them
// by comparing the provided MovieID with each booking's MovieID.
//
// Important behavior:
//   - Returns all bookings regardless of their status (confirmed, pending, canceled)
//   - Returns an empty slice when no bookings exist for the given movie
//   - The returned slice is a new slice; modifications won't affect the store
//   - The iteration order is not guaranteed (Go map iteration is random)
//
// Parameters:
//   - movieID: the identifier of the movie to filter bookings by
//
// Returns:
//   - []Booking: a new slice containing all bookings for the specified movie.
//     Returns an empty slice if no matching bookings are found.
func (store *MemoryStore) ListBookings(movieID string) []Booking {
	store.RLock()
	defer store.RUnlock()

	var result []Booking
	for _, booking := range store.bookings {
		if booking.MovieID == movieID {
			result = append(result, booking)
		}
	}

	return result
}
