package booking

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"concurrent_booking/internal/utils"
)

type handler struct {
	svc *Service
}

// NewHandler returns a new HTTP handler backed by the given booking service.
func NewHandler(svc *Service) *handler {
	return &handler{svc: svc}
}

type holdSeatRequest struct {
	UserID string `json:"user_id"`
}

// HoldSeat creates a temporary seat reservation.
//
// It returns the created booking session with its expiration time.
func (h *handler) HoldSeat(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")
	seatID := r.PathValue("seatID")

	var req holdSeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	session, err := h.svc.Book(Booking{
		UserID:  req.UserID,
		SeatID:  seatID,
		MovieID: movieID,
	})
	if err != nil {
		if errors.Is(err, ErrSeatAlreadyBooked) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to hold seat")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, holdResponse{
		SessionID: session.ID,
		MovieID:   session.MovieID,
		SeatID:    session.SeatID,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
	})
}

type holdResponse struct {
	SessionID string `json:"session_id"`
	MovieID   string `json:"movie_id"`
	SeatID    string `json:"seat_id"`
	ExpiresAt string `json:"expires_at"`
}

// ListSeats returns all currently booked seats for a movie.
func (h *handler) ListSeats(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")

	bookings := h.svc.ListBookings(movieID)

	seats := make([]seatInfo, 0, len(bookings))
	for _, booking := range bookings {
		seats = append(seats, seatInfo{
			SeatID:    booking.SeatID,
			UserID:    booking.UserID,
			Booked:    true,
			Confirmed: booking.Status == statusConfirmed,
		})
	}

	utils.WriteJSON(w, http.StatusOK, seats)
}

type seatInfo struct {
	SeatID    string `json:"seat_id"`
	UserID    string `json:"user_id"`
	Booked    bool   `json:"booked"`
	Confirmed bool   `json:"confirmed"`
}

// ConfirmSession confirms a booking session for the given user.
func (h *handler) ConfirmSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

	var req holdSeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	session, err := h.svc.ConfirmSeat(r.Context(), sessionID, req.UserID)
	if err != nil {
		writeSessionError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, sessionResponse{
		SessionID: session.ID,
		MovieID:   session.MovieID,
		SeatID:    session.SeatID,
		UserID:    session.UserID,
		Status:    session.Status,
	})
}

type sessionResponse struct {
	SessionID string `json:"session_id"`
	MovieID   string `json:"movie_id"`
	SeatID    string `json:"seat_id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// ReleaseSession releases a booking session for the given user.
func (h *handler) ReleaseSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

	var req holdSeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if err := h.svc.ReleaseSeat(r.Context(), sessionID, req.UserID); err != nil {
		writeSessionError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeSessionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrSessionNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrSessionNotOwned):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "failed to process session")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	utils.WriteJSON(w, status, map[string]string{
		"error": message,
	})
}
