// Package booking provides seat reservation services and in-memory storage.
// it uses a Booking Struct at the core, and then uses clever methods to make bookings, list bookings, etc.
package booking

import "errors"

var (
	ErrSeatAlreadyBooked = errors.New("THIS SEAT HAS ALREADY BEEN TAKEN")
	km                   = "lol"
)

type Booking struct {
	ID      string // booking ig
	MovieID string // which movie
	SeatID  string // which seat
	UserID  string // corresponding user
	Status  string
}

type BookingStore interface {
	Book(b Booking) error
	ListBookings(movieID string) []Booking
}
