package booking

import (
	"github.com/google/uuid"
	"time"
)

type BookingStatus string

const(
	BookingStatusCreated BookingStatus = "created"
	BookingStatusCancelled BookingStatus = "canceled"
	BookingStatusConfirmed BookingStatus = "confirmed"
)

type Booking struct {
	ID                   uuid.UUID     `json:"id"`
	EventID              uuid.UUID     `json:"event_id"`
	UserID               uuid.UUID     `json:"user_id"`
	EventName			 string		   `json:"event_name"`
	Count                int           `json:"count"`
	Price                float64       `json:"price"`
	Status               BookingStatus `json:"status"`
	CreatedAt            time.Time     `json:"created_at"`
	ExpiredAt            time.Time     `json:"expired_at"`
	TelegramNotification bool          `json:"telegram_notification"`
	TelegramRecepient string 		   `json:"telegram_recepient"`
	EmailNotification    bool          `json:"email_notification"`
	EmailRecepient string 		       `json:"email_recepient"`
}

func NewBooking(eventId, userId, TelegramRecepient, EmailRecepient, EventName string, telegramNotification, emailNotification bool, count int, expiredAt int, price float64) (*Booking, error) {
	eId, err := uuid.Parse(eventId)
	if err != nil {
		return nil, err
	}
	usId, err := uuid.Parse(userId)
	if err != nil {
		return nil, err
	}
	if count <= 0 {
		return nil, err
	}
	return &Booking{
		ID: uuid.New(),
		EventID: eId,
		UserID: usId,
		Count: count,
		Price: price*float64(count),
		CreatedAt: time.Now(),
		ExpiredAt: time.Now().Add(time.Minute*time.Duration(expiredAt)),
		Status: BookingStatusCreated,
		TelegramNotification: telegramNotification,
		EmailNotification: emailNotification,
		TelegramRecepient: TelegramRecepient,
		EmailRecepient: EmailRecepient,
		EventName: EventName,
	}, nil
}

func (b *Booking) StatusConfirmed (){	
	b.Status = BookingStatusConfirmed
}