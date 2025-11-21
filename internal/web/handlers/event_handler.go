package handlers

import (
	"time"
	"eventbooker/internal/domain/event"
	"eventbooker/internal/domain/booking"
	"eventbooker/internal/web/dto"
	wbgin "github.com/wb-go/wbf/ginext"
	"net/http"
)

type EventHandler struct {
	serviceEvent EventIFace
	serviceBooking BookingIFace
}

type EventIFace interface {
	CreateEvent(userid, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error)
	GetEvent(eventid string) (*event.Event, error)
}

type BookingIFace interface {
	CreateBooking(eventId, userId string, telegramNotification, emailNotification bool, count int) (*booking.Booking, error)
	ConfirmBooking(id string) error
}

func NewEventHandler(serviceEvent EventIFace, serviceBooking BookingIFace) *EventHandler {
	return &EventHandler{
		serviceEvent: serviceEvent,
		serviceBooking: serviceBooking,
	}
}

func (h *EventHandler) CreateEvent (ctx *wbgin.Context){
	var req dto.CreateEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}
	userId, _ := ctx.Params.Get("user_id")
	eventDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid date format"})
		return
	}
	event, err := h.serviceEvent.CreateEvent(userId, req.Name, req.Description, eventDate, req.BookingTTL, req.MaxCountPeople, req.Price)
	if  err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}
	res := dto.EventResponse{
		ID:             event.Id.String(),
		Name:           event.Name,
		Description:    event.Description,
		Date:           event.Date.Format(time.RFC3339),
		BookingTTL:     event.BookingTTL,
		MaxCountPeople: event.MaxCountPeople,
		Price:          event.Price,
	}
	ctx.JSON(http.StatusOK, res)
}

func (h *EventHandler) GetEvent (ctx *wbgin.Context){
	eventId, _ := ctx.Params.Get("event_id")
	event, err := h.serviceEvent.GetEvent(eventId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}
	var BookingResponses []dto.BookingResponse
	for _, booking := range event.Bookings {
		bookingResp := dto.BookingResponse{
			ID:                   booking.ID.String(),
			EventID:              booking.EventID.String(),
			UserID:               booking.UserID.String(),
			Status:               string(booking.Status),
			TelegramNotification: booking.TelegramNotification,
			EmailNotification:    booking.EmailNotification,
			Count:                booking.Count,
			ExpiredAt:            booking.ExpiredAt.Format(time.RFC3339),
			Price:                booking.Price,
		}
		BookingResponses = append(BookingResponses, bookingResp)
	}
	res := dto.EventResponse{
		ID:             event.Id.String(),
		Name:           event.Name,
		Description:    event.Description,
		Date:           event.Date.Format(time.RFC3339),
		BookingTTL:     event.BookingTTL,
		MaxCountPeople: event.MaxCountPeople,
		FreePlaces:     event.FreePlaces,
		Price:          event.Price,
		BookingResponses: BookingResponses,
	}
	ctx.JSON(http.StatusOK, res)
}

func (h *EventHandler) CreateBooking (ctx *wbgin.Context){
	var req dto.CreateBookingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}
	userId, _ := ctx.Params.Get("user_id")
	booking, err := h.serviceBooking.CreateBooking(req.EventID, userId, req.TelegramNotification, req.EmailNotification, req.Count)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}
	res := dto.BookingResponse{
		ID:                   booking.ID.String(),
		EventID:              booking.EventID.String(),
		UserID:               booking.UserID.String(),
		Status:               string(booking.Status),
		TelegramNotification: booking.TelegramNotification,
		EmailNotification:    booking.EmailNotification,
		Count:                booking.Count,
		ExpiredAt:            booking.ExpiredAt.Format(time.RFC3339),
		Price:                booking.Price,
	}
	ctx.JSON(http.StatusOK, res)
}

func (h *EventHandler) ConfirmBooking (ctx *wbgin.Context){
	bookingId, _ := ctx.Params.Get("booking_id")
	err := h.serviceBooking.ConfirmBooking(bookingId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, wbgin.H{"message": "booking confirmed"})
}
