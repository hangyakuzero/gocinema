// Package booking provides seat reservation services and in-memory storage.
// it uses a Booking Struct at the core, and then uses clever methods to make bookings, list bookings, etc.
package booking

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSeatAlreadyBooked        = errors.New("THIS SEAT HAS ALREADY BEEN TAKEN")
	ErrSessionDoesNotBelongUser = errors.New("session does not belong to user")
	ErrConfirmedCannotRelease   = errors.New("confirmed bookings cannot be released")
	km                          = "lol"
)

type Booking struct {
	ID        string // booking ig
	MovieID   string // which movie
	SeatID    string // which seat
	UserID    string // corresponding user
	Status    string
	ExpiresAt time.Time
}

type BookingStore interface {
	Book(b Booking) (Booking, error)
	ListBookings(movieID string) []Booking
	Confirm(ctx context.Context, sessionID string, userID string) (Booking, error)
	Release(ctx context.Context, sessionID string, userID string) error
}
