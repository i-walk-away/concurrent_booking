package booking

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func newTestRedisStore(t *testing.T) (*RedisStore, *miniredis.Miniredis) {
	t.Helper()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}

	t.Cleanup(server.Close)

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	t.Cleanup(func() {
		_ = client.Close()
	})

	return NewRedisStore(client), server
}

func TestRedisStore_Book(t *testing.T) {
	store, _ := newTestRedisStore(t)

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

func TestRedisStore_CannotBookSameSeatTwice(t *testing.T) {
	store, _ := newTestRedisStore(t)

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
	if err != ErrSeatAlreadyBooked {
		t.Fatalf("expected ErrSeatAlreadyBooked, got %v", err)
	}
}

func TestRedisStore_Confirm(t *testing.T) {
	store, _ := newTestRedisStore(t)

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

func TestRedisStore_CannotConfirmAnotherUsersSession(t *testing.T) {
	store, _ := newTestRedisStore(t)

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
	if err != ErrSessionNotOwned {
		t.Fatalf("expected ErrSessionNotOwned, got %v", err)
	}
}

func TestRedisStore_Release(t *testing.T) {
	store, _ := newTestRedisStore(t)

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

func TestRedisStore_HoldExpires(t *testing.T) {
	store, server := newTestRedisStore(t)

	_, err := store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-1",
	})
	if err != nil {
		t.Fatalf("book: %v", err)
	}

	server.FastForward(defaultHoldTTL + time.Second)

	_, err = store.Book(Booking{
		MovieID: "inception",
		SeatID:  "A1",
		UserID:  "user-2",
	})
	if err != nil {
		t.Fatalf("book after expiration: %v", err)
	}
}

// TestRedisStore_ConcurrentBookingExactlyOneWins verifies that concurrent
// attempts to book the same seat result in exactly one successful booking.
//
// The test uses miniredis, so it does not require a real Redis server.
func TestRedisStore_ConcurrentBookingExactlyOneWins(t *testing.T) {
	store, _ := newTestRedisStore(t)

	const numGoroutines = 100_000

	var (
		successes atomic.Int64
		failures  atomic.Int64
		wg        sync.WaitGroup
	)

	wg.Add(numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()

			_, err := store.Book(Booking{
				MovieID: "screen-1",
				SeatID:  "A1",
				UserID:  uuid.NewString(),
			})

			if err == nil {
				successes.Add(1)
				return
			}

			failures.Add(1)
		}()
	}

	wg.Wait()

	if got := successes.Load(); got != 1 {
		t.Errorf("expected exactly one success, got %d", got)
	}

	if got := failures.Load(); got != int64(numGoroutines-1) {
		t.Errorf(
			"expected %d failures, got %d",
			numGoroutines-1,
			got,
		)
	}
}
