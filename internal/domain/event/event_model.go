package event

import (
	"errors"
	"eventbooker/internal/domain/booking"
	"github.com/google/uuid"
	"time"
)

type Event struct {
	Id             uuid.UUID
	CreatorId      uuid.UUID
	Date           time.Time
	Name           string
	Description    string
	MaxCountPeople int
	FreePlaces     int
	Price          float64
	BookingTTL     int
	Bookings       []*booking.Booking
}

func NewEvent(Creator, Name, Description string, Date time.Time, bookingTTL, MaxCountPeople int, Price float64) (*Event, error) {
	creatorUid, err := uuid.Parse(Creator)
	if err != nil {
		return nil, err
	}
	if MaxCountPeople <= 0 {
		return nil, errors.New("count of people should be bigger than 0")
	}
	if Price < 0 {
		return nil, errors.New("price must be bigger or equal 0")
	}

	return &Event{
		Id:             uuid.New(),
		CreatorId:      creatorUid,
		Date:           Date,
		Name:           Name,
		Description:    Description,
		MaxCountPeople: MaxCountPeople,
		FreePlaces:     MaxCountPeople,
		BookingTTL:     bookingTTL,
		Price:          Price,
		Bookings:       []*booking.Booking{},
	}, nil
}
