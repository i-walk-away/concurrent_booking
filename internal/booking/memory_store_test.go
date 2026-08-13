package booking

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestInMemoryStore_Book(t *testing.T) {
	store := NewInMemoryStore()

	booking, err := store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("book: %v", err)
	}

	if booking.ID == "" {
		t.Fatal("expected booking ID to be generated")
	}

	if booking.Status != statusHeld {
		t.Fatalf("expected status %q, got %q", statusHeld, booking.Status)
	}

	if booking.ExpiresAt.IsZero() {
		t.Fatal("expected expiration time to be set")
	}
}

func TestInMemoryStore_CannotBookSameSeatTwice(t *testing.T) {
	store := NewInMemoryStore()

	_, err := store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("first booking: %v", err)
	}

	_, err = store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-2",
	})
	if !errors.Is(err, ErrSeatAlreadyBooked) {
		t.Fatalf("expected ErrSeatAlreadyBooked, got %v", err)
	}
}

func TestInMemoryStore_DifferentMoviesCanUseSameSeat(t *testing.T) {
	store := NewInMemoryStore()

	_, err := store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("first booking: %v", err)
	}

	_, err = store.Book(Booking{
		MovieID: "dune",
		SeatID:  "A1",
		UserID:  "user-2",
	})
	if err != nil {
		t.Fatalf("second booking: %v", err)
	}
}

func TestInMemoryStore_Confirm(t *testing.T) {
	store := NewInMemoryStore()

	booking, err := store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("book: %v", err)
	}

	confirmed, err := store.Confirm(
		context.Background(),
		booking.ID,
		"user-1",
	)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}

	if confirmed.Status != statusConfirmed {
		t.Fatalf("expected status %q, got %q", statusConfirmed, confirmed.Status)
	}

	if !confirmed.ExpiresAt.IsZero() {
		t.Fatal("confirmed booking must not expire")
	}
}

func TestInMemoryStore_CannotConfirmAnotherUsersSession(t *testing.T) {
	store := NewInMemoryStore()

	booking, err := store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("book: %v", err)
	}

	_, err = store.Confirm(
		context.Background(),
		booking.ID,
		"user-2",
	)
	if !errors.Is(err, ErrSessionNotOwned) {
		t.Fatalf("expected ErrSessionNotOwned, got %v", err)
	}
}

func TestInMemoryStore_Release(t *testing.T) {
	store := NewInMemoryStore()

	booking, err := store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("book: %v", err)
	}

	if err := store.Release(
		context.Background(),
		booking.ID,
		"user-1",
	); err != nil {
		t.Fatalf("release: %v", err)
	}

	_, err = store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-2",
	})
	if err != nil {
		t.Fatalf("book after release: %v", err)
	}
}

func TestInMemoryStore_ExpiredHoldCanBeBookedAgain(t *testing.T) {
	store := NewInMemoryStore()

	booking, err := store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("book: %v", err)
	}

	store.mu.Lock()
	booking.ExpiresAt = time.Now().Add(-time.Second)
	store.bookings[booking.ID] = booking
	store.mu.Unlock()

	_, err = store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-2",
	})
	if err != nil {
		t.Fatalf("book after expiration: %v", err)
	}
}
