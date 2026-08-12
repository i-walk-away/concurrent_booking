package booking

// Booking represents a confirmed seat reservation for a specific movie screening.
// It contains all the necessary information to identify the reservation and its
// current state in the booking lifecycle.
//
// Fields:
//   - ID: unique identifier for the booking record
//   - MovieID: identifier of the movie being booked
//   - SeatID: unique identifier of the seat being reserved
//   - UserID: identifier of the user making the reservation
//   - Status: current state of the booking (e.g., "confirmed", "pending", "canceled")
type Booking struct {
	ID      string
	MovieID string
	SeatID  string
	UserID  string
	Status  string
}

// BookingStore defines the contract for storing and retrieving booking records.
// Implementations of this interface can use different storage backends such as
// in-memory maps, relational databases, or distributed caches.
//
// The interface provides two primary operations:
//   - Book: atomically reserves a seat for a user
//   - ListBookings: retrieves all bookings for a specific movie
type BookingStore interface {
	// Book attempts to reserve a seat for a user.
	// It returns an error if the seat is already taken.
	//
	// Parameters:
	//   - booking: Booking to be stored
	//
	// Returns:
	//   - error: nil if booking was successful, ErrSeatAlreadyTaken if seat is occupied
	Book(booking Booking) error

	// ListBookings returns all bookings for a specific movie.
	// If no bookings exist for the movie, an empty slice is returned.
	//
	// Parameters:
	//   - movieID: identifier of the movie to filter bookings by
	//
	// Returns:
	//   - []Booking: slice containing all bookings for the specified movie
	ListBookings(movieID string) []Booking
}
