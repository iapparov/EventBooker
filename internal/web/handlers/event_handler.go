package handlers

import (
	"eventbooker/internal/domain/booking"
	"eventbooker/internal/domain/event"
	"eventbooker/internal/web/dto"
	"fmt"
	wbgin "github.com/wb-go/wbf/ginext"
	"net/http"
	"time"
)

type EventHandler struct {
	serviceEvent   EventIFace
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
		serviceEvent:   serviceEvent,
		serviceBooking: serviceBooking,
	}
}

// CreateEvent godoc
// @Summary      Create a new event
// @Description  Create a new event for the authenticated user
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateEventRequest  true  "Event info"
// @Success      200   {object}  dto.EventResponse
// @Failure      400   {object}  map[string]string  "Invalid request"
// @Failure      401   {object}  map[string]string  "Unauthorized"
// @Failure      500   {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /events [post]
func (h *EventHandler) CreateEvent(ctx *wbgin.Context) {
	var req dto.CreateEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}
	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, wbgin.H{"error": "user not found in context"})
		return
	}
	eventDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid date format"})
		return
	}
	event, err := h.serviceEvent.CreateEvent(userId.(string), req.Name, req.Description, eventDate, req.BookingTTL, req.MaxCountPeople, req.Price)
	if err != nil {
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

// GetEvent godoc
// @Summary      Get event by ID
// @Description  Retrieve an event and its bookings by event ID
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Event ID"
// @Success      200   {object}  dto.EventResponse
// @Failure      500   {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /events/{id} [get]
func (h *EventHandler) GetEvent(ctx *wbgin.Context) {
	eventId, _ := ctx.Params.Get("id")
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
		ID:               event.Id.String(),
		Name:             event.Name,
		Description:      event.Description,
		Date:             event.Date.Format(time.RFC3339),
		BookingTTL:       event.BookingTTL,
		MaxCountPeople:   event.MaxCountPeople,
		FreePlaces:       event.FreePlaces,
		Price:            event.Price,
		BookingResponses: BookingResponses,
	}
	ctx.JSON(http.StatusOK, res)
}

// CreateBooking godoc
// @Summary      Create a booking for an event
// @Description  Book a number of seats for the authenticated user
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateBookingRequest  true  "Booking info"
// @Success      200   {object}  dto.BookingResponse
// @Failure      400   {object}  map[string]string  "Invalid request"
// @Failure      401   {object}  map[string]string  "Unauthorized"
// @Failure      500   {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /bookings [post]
func (h *EventHandler) CreateBooking(ctx *wbgin.Context) {
	var req dto.CreateBookingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}
	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, wbgin.H{"error": "user not found in context"})
		return
	}
	booking, err := h.serviceBooking.CreateBooking(req.EventID, userId.(string), req.TelegramNotification, req.EmailNotification, req.Count)
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

// ConfirmBooking godoc
// @Summary      Confirm a booking
// @Description  Confirm a previously created booking by booking ID
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "Booking ID"
// @Success      200   {object}  map[string]string  "Booking confirmed"
// @Failure      500   {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /bookings/{id}/confirm [post]
func (h *EventHandler) ConfirmBooking(ctx *wbgin.Context) {
	bookingId, _ := ctx.Params.Get("id")
	fmt.Println(bookingId)
	err := h.serviceBooking.ConfirmBooking(bookingId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, wbgin.H{"message": "booking confirmed"})
}
