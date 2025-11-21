package dto

type CreateEventRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description" binding:"required"`
	Date          string  `json:"date" binding:"required,datetime=2006-01-02T15:04:05Z07:00"`
	BookingTTL    int     `json:"booking_ttl" binding:"required,min=1"`
	MaxCountPeople int    `json:"max_count_people" binding:"required,min=1"`
	Price         float64 `json:"price" binding:"required,min=0"`
}

type EventResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Date          string  `json:"date"`
	BookingTTL    int     `json:"booking_ttl"`
	MaxCountPeople int    `json:"max_count_people"`
	FreePlaces    int     `json:"free_places,omitempty"`
	Price         float64 `json:"price"`
	BookingResponses []BookingResponse `json:"bookings,omitempty"`
}

type CreateBookingRequest struct {
	EventID              string `json:"event_id" binding:"required"`
	TelegramNotification bool `json:"telegram_notification"`
	EmailNotification    bool `json:"email_notification"`
	Count                int  `json:"count" binding:"required,min=1"`
}

type BookingResponse struct {
	ID                   string  `json:"id"`
	EventID              string  `json:"event_id"`
	UserID               string  `json:"user_id"`
	Status			     string  `json:"status"`
	TelegramNotification bool    `json:"telegram_notification"`
	EmailNotification    bool    `json:"email_notification"`
	Count                int     `json:"count"`
	ExpiredAt            string  `json:"expired_at"`
	Price                float64 `json:"price"`
}