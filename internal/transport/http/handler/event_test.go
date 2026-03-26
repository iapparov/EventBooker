package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eventbooker/internal/domain/booking"
	"eventbooker/internal/domain/event"
	"eventbooker/internal/transport/http/dto"
	"eventbooker/internal/transport/http/handler"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockEventService struct {
	CreateFn func(ctx context.Context, userID, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error)
	GetFn    func(ctx context.Context, eventID string) (*event.Event, error)
}

func (m *mockEventService) Create(ctx context.Context, userID, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error) {
	return m.CreateFn(ctx, userID, name, description, date, bookingTTL, maxCountPeople, price)
}
func (m *mockEventService) Get(ctx context.Context, eventID string) (*event.Event, error) {
	return m.GetFn(ctx, eventID)
}

type mockBookingService struct {
	CreateFn  func(ctx context.Context, eventID, userID string, tg, email bool, count int) (*booking.Booking, error)
	ConfirmFn func(ctx context.Context, id string) error
}

func (m *mockBookingService) Create(ctx context.Context, eventID, userID string, tg, email bool, count int) (*booking.Booking, error) {
	return m.CreateFn(ctx, eventID, userID, tg, email, count)
}
func (m *mockBookingService) Confirm(ctx context.Context, id string) error {
	return m.ConfirmFn(ctx, id)
}

func performRequest(hf func(*gin.Context), method, path string, body any, userID string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	if userID != "" {
		c.Set("userId", userID)
	}
	hf(c)
	return w
}

func TestEventHandler_CreateEvent_Success(t *testing.T) {
	mock := &mockEventService{
		CreateFn: func(ctx context.Context, userID, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error) {
			return &event.Event{
				ID: uuid.New(), Name: name, Description: description, Date: date,
				MaxCountPeople: maxCountPeople, BookingTTL: bookingTTL, Price: price,
			}, nil
		},
	}
	h := handler.NewEventHandler(mock, nil)
	req := dto.CreateEventRequest{
		Name: "Test Event", Description: "desc",
		Date:           time.Now().Add(time.Hour).Format(time.RFC3339),
		MaxCountPeople: 10, BookingTTL: 5, Price: 100,
	}
	w := performRequest(h.CreateEvent, "POST", "/events", req, "user-123")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestEventHandler_CreateEvent_InvalidJSON(t *testing.T) {
	h := handler.NewEventHandler(&mockEventService{}, nil)
	w := performRequest(h.CreateEvent, "POST", "/events", "{bad json", "user-123")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_CreateEvent_NoUser(t *testing.T) {
	h := handler.NewEventHandler(&mockEventService{}, nil)
	req := dto.CreateEventRequest{
		Name: "Test Event", Description: "desc",
		Date:           time.Now().Add(time.Hour).Format(time.RFC3339),
		MaxCountPeople: 10, BookingTTL: 5, Price: 100,
	}
	w := performRequest(h.CreateEvent, "POST", "/events", req, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestEventHandler_CreateEvent_InvalidDate(t *testing.T) {
	h := handler.NewEventHandler(&mockEventService{}, nil)
	req := dto.CreateEventRequest{
		Name: "Test Event", Description: "desc", Date: "invalid-date",
		MaxCountPeople: 10, BookingTTL: 5, Price: 100,
	}
	w := performRequest(h.CreateEvent, "POST", "/events", req, "user-123")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_CreateEvent_ServiceError(t *testing.T) {
	mock := &mockEventService{
		CreateFn: func(ctx context.Context, userID, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error) {
			return nil, errors.New("service error")
		},
	}
	h := handler.NewEventHandler(mock, nil)
	req := dto.CreateEventRequest{
		Name: "Test Event", Description: "desc",
		Date:           time.Now().Add(time.Hour).Format(time.RFC3339),
		MaxCountPeople: 10, BookingTTL: 5, Price: 100,
	}
	w := performRequest(h.CreateEvent, "POST", "/events", req, "user-123")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestEventHandler_GetEvent_Success(t *testing.T) {
	ev := &event.Event{
		ID: uuid.New(), Name: "Test", FreePlaces: 5, Date: time.Now(),
		Bookings: []*booking.Booking{}, BookingTTL: 10, MaxCountPeople: 50,
	}
	mock := &mockEventService{GetFn: func(ctx context.Context, eventID string) (*event.Event, error) { return ev, nil }}
	h := handler.NewEventHandler(mock, nil)
	w := performRequest(h.GetEvent, "GET", "/events/"+ev.ID.String(), nil, "u1")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestEventHandler_GetEvent_Error(t *testing.T) {
	mock := &mockEventService{GetFn: func(ctx context.Context, eventID string) (*event.Event, error) { return nil, errors.New("not found") }}
	h := handler.NewEventHandler(mock, nil)
	w := performRequest(h.GetEvent, "GET", "/events/123", nil, "u1")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestEventHandler_GetEvent_WithBookings(t *testing.T) {
	b := &booking.Booking{
		ID: uuid.New(), EventID: uuid.New(), UserID: uuid.New(),
		Status: booking.StatusCreated, Count: 2, Price: 100,
		ExpiredAt: time.Now().Add(10 * time.Minute),
	}
	ev := &event.Event{
		ID: uuid.New(), Name: "Test Event", FreePlaces: 5, Date: time.Now(),
		Bookings: []*booking.Booking{b}, BookingTTL: 10, MaxCountPeople: 50,
	}
	mock := &mockEventService{GetFn: func(ctx context.Context, eventID string) (*event.Event, error) { return ev, nil }}
	h := handler.NewEventHandler(mock, nil)
	w := performRequest(h.GetEvent, "GET", "/events/"+ev.ID.String(), nil, "u1")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestEventHandler_CreateBooking_Success(t *testing.T) {
	mock := &mockBookingService{
		CreateFn: func(ctx context.Context, eventID, userID string, tg, email bool, count int) (*booking.Booking, error) {
			return &booking.Booking{
				ID: uuid.New(), EventID: uuid.New(), UserID: uuid.New(),
				Status: booking.StatusCreated, Count: count,
				ExpiredAt: time.Now().Add(10 * time.Minute), Price: 50,
			}, nil
		},
	}
	h := handler.NewEventHandler(nil, mock)
	req := dto.CreateBookingRequest{EventID: uuid.New().String(), Count: 1, TelegramNotification: true}
	w := performRequest(h.CreateBooking, "POST", "/book", req, "user123")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestEventHandler_CreateBooking_InvalidJSON(t *testing.T) {
	h := handler.NewEventHandler(nil, &mockBookingService{})
	w := performRequest(h.CreateBooking, "POST", "/book", "{bad json", "u")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_CreateBooking_NoUser(t *testing.T) {
	h := handler.NewEventHandler(nil, &mockBookingService{})
	req := dto.CreateBookingRequest{EventID: uuid.New().String(), Count: 1, TelegramNotification: true}
	w := performRequest(h.CreateBooking, "POST", "/book", req, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestEventHandler_CreateBooking_ServiceError(t *testing.T) {
	mock := &mockBookingService{
		CreateFn: func(ctx context.Context, eventID, userID string, tg, email bool, count int) (*booking.Booking, error) {
			return nil, errors.New("service error")
		},
	}
	h := handler.NewEventHandler(nil, mock)
	req := dto.CreateBookingRequest{EventID: uuid.New().String(), Count: 1, TelegramNotification: true}
	w := performRequest(h.CreateBooking, "POST", "/book", req, "user123")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestEventHandler_ConfirmBooking_Success(t *testing.T) {
	mock := &mockBookingService{ConfirmFn: func(ctx context.Context, id string) error { return nil }}
	h := handler.NewEventHandler(nil, mock)
	w := performRequest(h.ConfirmBooking, "POST", "/confirm/12345", nil, "user")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestEventHandler_ConfirmBooking_Error(t *testing.T) {
	mock := &mockBookingService{ConfirmFn: func(ctx context.Context, id string) error { return errors.New("failed") }}
	h := handler.NewEventHandler(nil, mock)
	w := performRequest(h.ConfirmBooking, "POST", "/confirm/77", nil, "user")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
