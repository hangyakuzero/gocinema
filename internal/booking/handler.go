package booking

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/hangyakuzero/gocinema/internal/utils"
)

type handler struct {
	svc *Service
}

type seatInfo struct {
	SeatID    string `json:"seat_id"`
	UserID    string `json:"user_id"`
	Booked    bool   `json:"booked"`
	Confirmed bool   `json:"confirmed"`
}

type sessionRequest struct {
	UserID string `json:"user_id"`
}

func NewHandler(svc *Service) *handler {
	return &handler{svc}
}

func (h *handler) HoldSeat(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")
	seatID := r.PathValue("seatID")

	type holdResquest struct {
		UserID string `json:"user_id"`
	}

	var req holdResquest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return

	}
	data := Booking{
		UserID:  req.UserID,
		SeatID:  seatID,
		MovieID: movieID,
	}

	session, err := h.svc.Book(data)
	if err != nil {
		log.Println(err)
		if errors.Is(err, ErrSeatAlreadyBooked) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to hold seat")
		return
	}

	type holdResponse struct {
		SessionID string `json:"session_id"`
		MovieID   string `json:"movie_id"`
		SeatID    string `json:"seat_id"`
		ExpiresAt string `json:"expires_at"`
	}

	utils.WriteJSON(w, http.StatusCreated, holdResponse{
		SeatID:    seatID,
		MovieID:   session.MovieID,
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *handler) ListSeats(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")
	bookings := h.svc.ListBookings(movieID)
	seats := make([]seatInfo, 0, len(bookings))

	for _, b := range bookings {
		seats = append(seats, seatInfo{
			SeatID:    b.SeatID,
			UserID:    b.UserID,
			Booked:    true,
			Confirmed: b.Status == "confirmed",
		})
	}
	utils.WriteJSON(w, http.StatusOK, seats)
}

func (h *handler) ConfirmSeat(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

	var req sessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	booking, err := h.svc.Confirm(r.Context(), sessionID, req.UserID)
	if err != nil {
		log.Println(err)
		status := http.StatusBadRequest
		if errors.Is(err, ErrSessionDoesNotBelongUser) {
			status = http.StatusForbidden
		}
		writeError(w, status, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, booking)
}

func (h *handler) ReleaseSeat(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

	var req sessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.Release(r.Context(), sessionID, req.UserID); err != nil {
		log.Println(err)
		status := http.StatusBadRequest
		if errors.Is(err, ErrSessionDoesNotBelongUser) {
			status = http.StatusForbidden
		}
		writeError(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, status int, message string) {
	utils.WriteJSON(w, status, map[string]string{"error": message})
}
