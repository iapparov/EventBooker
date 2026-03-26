package event

import (
	"errors"
	"time"

	"eventbooker/internal/domain/booking"

	"github.com/google/uuid"
)

// Event is the domain model for an event.
type Event struct {
	ID             uuid.UUID
	CreatorID      uuid.UUID
	Date           time.Time
	Name           string
	Description    string
	MaxCountPeople int
	FreePlaces     int
	Price          float64
	BookingTTL     int
	Bookings       []*booking.Booking
}

// New creates a new Event with validation.
func New(creator, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*Event, error) {
	creatorUID, err := uuid.Parse(creator)
	if err != nil {
		return nil, err
	}

	if maxCountPeople <= 0 {
		return nil, errors.New("count of people should be bigger than 0")
	}

	if price < 0 {
		return nil, errors.New("price must be bigger or equal 0")
	}

	return &Event{
		ID:             uuid.New(),
		CreatorID:      creatorUID,
		Date:           date,
		Name:           name,
		Description:    description,
		MaxCountPeople: maxCountPeople,
		FreePlaces:     maxCountPeople,
		BookingTTL:     bookingTTL,
		Price:          price,
		Bookings:       []*booking.Booking{},
	}, nil
}
