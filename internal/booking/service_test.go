package booking

import (
	"context"
	"errors"
	"testing"
)

func TestServiceBook(t *testing.T) {
	store := NewInMemoryStore()
	svc := NewService(store)

	booking, err := svc.Book(Booking{
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

	if booking.MovieID != "inception" {
		t.Errorf("expected movie ID %q, got %q", "inception", booking.MovieID)
	}

	if booking.SeatID != "A1" {
		t.Errorf("expected seat ID %q, got %q", "A1", booking.SeatID)
	}

	if booking.UserID != "user-1" {
		t.Errorf("expected user ID %q, got %q", "user-1", booking.UserID)
	}

	if booking.Status != statusHeld {
		t.Errorf("expected status %q, got %q", statusHeld, booking.Status)
	}
}

func TestServiceListBookings(t *testing.T) {
	store := NewInMemoryStore()
	svc := NewService(store)

	_, err := svc.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("book A1: %v", err)
	}

	_, err = svc.Book(Booking{
		MovieID: "inception",
		SeatID:  "A2",
		UserID:  "user-2",
	})
	if err != nil {
		t.Fatalf("book A2: %v", err)
	}

	_, err = svc.Book(Booking{
		MovieID: "dune",
		SeatID:  "A1",
		UserID:  "user-3",
	})
	if err != nil {
		t.Fatalf("book Dune A1: %v", err)
	}

	bookings := svc.ListBookings("inception")

	if len(bookings) != 2 {
		t.Fatalf("expected 2 bookings, got %d", len(bookings))
	}
}

func TestServiceConfirmSeat(t *testing.T) {
	store := NewInMemoryStore()
	svc := NewService(store)

	booking, err := svc.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("book: %v", err)
	}

	confirmed, err := svc.ConfirmSeat(
		context.Background(),
		booking.ID,
		"user-1",
	)
	if err != nil {
		t.Fatalf("confirm seat: %v", err)
	}

	if confirmed.Status != statusConfirmed {
		t.Errorf(
			"expected status %q, got %q",
			statusConfirmed,
			confirmed.Status,
		)
	}
}

func TestServiceConfirmSeatRejectsAnotherUser(t *testing.T) {
	store := NewInMemoryStore()
	svc := NewService(store)

	booking, err := svc.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("book: %v", err)
	}

	_, err = svc.ConfirmSeat(
		context.Background(),
		booking.ID,
		"user-2",
	)
	if !errors.Is(err, ErrSessionNotOwned) {
		t.Fatalf("expected ErrSessionNotOwned, got %v", err)
	}
}

func TestServiceReleaseSeat(t *testing.T) {
	store := NewInMemoryStore()
	svc := NewService(store)

	booking, err := svc.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("book: %v", err)
	}

	if err := svc.ReleaseSeat(
		context.Background(),
		booking.ID,
		"user-1",
	); err != nil {
		t.Fatalf("release seat: %v", err)
	}

	_, err = svc.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-2",
	})
	if err != nil {
		t.Fatalf("book after release: %v", err)
	}
}
