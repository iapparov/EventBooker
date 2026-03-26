package handler

import (
	"context"
	"net/http"
	"time"

	"eventbooker/internal/domain/booking"
	"eventbooker/internal/domain/event"
	"eventbooker/internal/transport/http/dto"

	wbgin "github.com/wb-go/wbf/ginext"
)

// EventServicer defines the event service interface used by EventHandler.
type EventServicer interface {
	Create(ctx context.Context, userID, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error)
	Get(ctx context.Context, eventID string) (*event.Event, error)
}

// BookingServicer defines the booking service interface used by EventHandler.
type BookingServicer interface {
	Create(ctx context.Context, eventID, userID string, telegramNotification, emailNotification bool, count int) (*booking.Booking, error)
	Confirm(ctx context.Context, id string) error
}

// EventHandler handles HTTP requests for events and bookings.
type EventHandler struct {
	events   EventServicer
	bookings BookingServicer
}

// NewEventHandler creates a new EventHandler.
func NewEventHandler(events EventServicer, bookings BookingServicer) *EventHandler {
	return &EventHandler{
		events:   events,
		bookings: bookings,
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

	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, wbgin.H{"error": "user not found in context"})
		return
	}

	eventDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid date format"})
		return
	}

	ev, err := h.events.Create(ctx.Request.Context(), userID.(string), req.Name, req.Description, eventDate, req.BookingTTL, req.MaxCountPeople, req.Price)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.EventResponse{
		ID:             ev.ID.String(),
		Name:           ev.Name,
		Description:    ev.Description,
		Date:           ev.Date.Format(time.RFC3339),
		BookingTTL:     ev.BookingTTL,
		MaxCountPeople: ev.MaxCountPeople,
		Price:          ev.Price,
	})
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
	eventID, _ := ctx.Params.Get("id")

	ev, err := h.events.Get(ctx.Request.Context(), eventID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}

	var bookingResponses []dto.BookingResponse
	for _, b := range ev.Bookings {
		bookingResponses = append(bookingResponses, dto.BookingResponse{
			ID:                   b.ID.String(),
			EventID:              b.EventID.String(),
			UserID:               b.UserID.String(),
			Status:               string(b.Status),
			TelegramNotification: b.TelegramNotification,
			EmailNotification:    b.EmailNotification,
			Count:                b.Count,
			ExpiredAt:            b.ExpiredAt.Format(time.RFC3339),
			Price:                b.Price,
		})
	}

	ctx.JSON(http.StatusOK, dto.EventResponse{
		ID:               ev.ID.String(),
		Name:             ev.Name,
		Description:      ev.Description,
		Date:             ev.Date.Format(time.RFC3339),
		BookingTTL:       ev.BookingTTL,
		MaxCountPeople:   ev.MaxCountPeople,
		FreePlaces:       ev.FreePlaces,
		Price:            ev.Price,
		BookingResponses: bookingResponses,
	})
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

	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, wbgin.H{"error": "user not found in context"})
		return
	}

	b, err := h.bookings.Create(ctx.Request.Context(), req.EventID, userID.(string), req.TelegramNotification, req.EmailNotification, req.Count)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.BookingResponse{
		ID:                   b.ID.String(),
		EventID:              b.EventID.String(),
		UserID:               b.UserID.String(),
		Status:               string(b.Status),
		TelegramNotification: b.TelegramNotification,
		EmailNotification:    b.EmailNotification,
		Count:                b.Count,
		ExpiredAt:            b.ExpiredAt.Format(time.RFC3339),
		Price:                b.Price,
	})
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
	bookingID, _ := ctx.Params.Get("id")

	if err := h.bookings.Confirm(ctx.Request.Context(), bookingID); err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, wbgin.H{"message": "booking confirmed"})
}
