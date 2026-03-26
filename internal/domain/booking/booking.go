package booking

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Status represents the state of a booking.
type Status string

const (
	StatusCreated   Status = "created"
	StatusCancelled Status = "canceled"
	StatusConfirmed Status = "confirmed"
)

// Booking is the domain model for a reservation.
type Booking struct {
	ID                   uuid.UUID `json:"id"`
	EventID              uuid.UUID `json:"event_id"`
	UserID               uuid.UUID `json:"user_id"`
	EventName            string    `json:"event_name"`
	Count                int       `json:"count"`
	Price                float64   `json:"price"`
	Status               Status    `json:"status"`
	CreatedAt            time.Time `json:"created_at"`
	ExpiredAt            time.Time `json:"expired_at"`
	TelegramNotification bool      `json:"telegram_notification"`
	TelegramRecepient    string    `json:"telegram_recepient"`
	EmailNotification    bool      `json:"email_notification"`
	EmailRecepient       string    `json:"email_recepient"`
}

// New creates a new Booking with validation.
func New(eventID, userID, telegramRecepient, emailRecepient, eventName string, telegramNotification, emailNotification bool, count int, expiredAtMinutes int, price float64) (*Booking, error) {
	eID, err := uuid.Parse(eventID)
	if err != nil {
		return nil, err
	}

	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	if count <= 0 {
		return nil, errors.New("count should be bigger than 0")
	}

	return &Booking{
		ID:                   uuid.New(),
		EventID:              eID,
		UserID:               uID,
		Count:                count,
		Price:                price * float64(count),
		CreatedAt:            time.Now(),
		ExpiredAt:            time.Now().Add(time.Minute * time.Duration(expiredAtMinutes)),
		Status:               StatusCreated,
		TelegramNotification: telegramNotification,
		EmailNotification:    emailNotification,
		TelegramRecepient:    telegramRecepient,
		EmailRecepient:       emailRecepient,
		EventName:            eventName,
	}, nil
}

// Confirm sets the booking status to confirmed.
func (b *Booking) Confirm() {
	b.Status = StatusConfirmed
}
