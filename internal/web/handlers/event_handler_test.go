package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"eventbooker/internal/domain/booking"
	"eventbooker/internal/domain/event"
	"eventbooker/internal/web/dto"
	"eventbooker/internal/web/handlers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ---------------- MOCKS --------------------

type MockEventService struct {
	CreateEventFn func(userid, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error)
	GetEventFn    func(eventid string) (*event.Event, error)
}

func (m *MockEventService) CreateEvent(userid, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error) {
	return m.CreateEventFn(userid, name, description, date, bookingTTL, maxCountPeople, price)
}

func (m *MockEventService) GetEvent(eventid string) (*event.Event, error) {
	return m.GetEventFn(eventid)
}

type MockBookingService struct {
	CreateBookingFn  func(eventId, userId string, telegramNotification, emailNotification bool, count int) (*booking.Booking, error)
	ConfirmBookingFn func(id string) error
}

func (m *MockBookingService) CreateBooking(eventId, userId string, telegramNotification, emailNotification bool, count int) (*booking.Booking, error) {
	return m.CreateBookingFn(eventId, userId, telegramNotification, emailNotification, count)
}

func (m *MockBookingService) ConfirmBooking(id string) error {
	return m.ConfirmBookingFn(id)
}

// ----------- UTIL --------------

func performRequest(hf func(*gin.Context), method, path string, body any, userId string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}

	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	if userId != "" {
		c.Set("userId", userId)
	}

	hf(c) // передаем обычный *gin.Context

	return w
}

// ---------------- TESTS -----------------------------

func TestEventHandler_CreateEvent_Success(t *testing.T) {
	mockEvent := &MockEventService{
		CreateEventFn: func(userid, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error) {
			return &event.Event{
				Id:             uuid.New(),
				Name:           name,
				Description:    description,
				Date:           date,
				MaxCountPeople: maxCountPeople,
				BookingTTL:     bookingTTL,
				Price:          price,
			}, nil
		},
	}

	h := handlers.NewEventHandler(mockEvent, nil)

	req := dto.CreateEventRequest{
		Name:           "Test Event",
		Description:    "desc",
		Date:           time.Now().Add(time.Hour).Format(time.RFC3339),
		MaxCountPeople: 10,
		BookingTTL:     5,
		Price:          100,
	}

	w := performRequest(h.CreateEvent, "POST", "/events", req, "user-123")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestEventHandler_CreateEvent_InvalidJSON(t *testing.T) {
	h := handlers.NewEventHandler(&MockEventService{}, nil)

	w := performRequest(h.CreateEvent, "POST", "/events", "{bad json", "user-123")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_CreateEvent_NoUser(t *testing.T) {
	h := handlers.NewEventHandler(&MockEventService{}, nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/events", func(c *gin.Context) { h.CreateEvent(c) })

	// передаём валидный JSON, но без userId
	req := dto.CreateEventRequest{
		Name:           "Test Event",
		Description:    "desc",
		Date:           time.Now().Add(time.Hour).Format(time.RFC3339),
		MaxCountPeople: 10,
		BookingTTL:     5,
		Price:          100,
	}

	w := performRequest(func(c *gin.Context) {
		h.CreateEvent(c)
	}, "POST", "/events", req, "") // пустой userId

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestEventHandler_GetEvent_Success(t *testing.T) {
	ev := &event.Event{
		Id:             uuid.New(),
		Name:           "Test",
		FreePlaces:     5,
		Date:           time.Now(),
		Bookings:       []*booking.Booking{},
		BookingTTL:     10,
		MaxCountPeople: 50,
	}

	mockEvent := &MockEventService{
		GetEventFn: func(eventid string) (*event.Event, error) {
			return ev, nil
		},
	}

	h := handlers.NewEventHandler(mockEvent, nil)

	w := performRequest(h.GetEvent, "GET", "/events/"+ev.Id.String(), nil, "u1")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestEventHandler_GetEvent_Error(t *testing.T) {
	mockEvent := &MockEventService{
		GetEventFn: func(eventid string) (*event.Event, error) {
			return nil, errors.New("not found")
		},
	}

	h := handlers.NewEventHandler(mockEvent, nil)

	w := performRequest(h.GetEvent, "GET", "/events/123", nil, "u1")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestEventHandler_CreateBooking_Success(t *testing.T) {
	mockBooking := &MockBookingService{
		CreateBookingFn: func(eventId, userId string, tg, mail bool, count int) (*booking.Booking, error) {
			return &booking.Booking{
				ID:        uuid.New(),
				EventID:   uuid.New(),
				UserID:    uuid.New(),
				Status:    booking.BookingStatusCreated,
				Count:     count,
				ExpiredAt: time.Now().Add(10 * time.Minute),
				Price:     50,
			}, nil
		},
	}

	h := handlers.NewEventHandler(nil, mockBooking)

	req := dto.CreateBookingRequest{
		EventID:              uuid.New().String(),
		Count:                1,
		TelegramNotification: true,
		EmailNotification:    false,
	}

	w := performRequest(h.CreateBooking, "POST", "/book", req, "user123")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestEventHandler_CreateBooking_InvalidJSON(t *testing.T) {
	h := handlers.NewEventHandler(nil, &MockBookingService{})

	w := performRequest(h.CreateBooking, "POST", "/book", "{bad json", "u")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_ConfirmBooking_Success(t *testing.T) {
	mockBooking := &MockBookingService{
		ConfirmBookingFn: func(id string) error { return nil },
	}

	h := handlers.NewEventHandler(nil, mockBooking)

	w := performRequest(h.ConfirmBooking, "POST", "/confirm/12345", nil, "user")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestEventHandler_ConfirmBooking_Error(t *testing.T) {
	mockBooking := &MockBookingService{
		ConfirmBookingFn: func(id string) error { return errors.New("failed") },
	}

	h := handlers.NewEventHandler(nil, mockBooking)

	w := performRequest(h.ConfirmBooking, "POST", "/confirm/77", nil, "user")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestEventHandler_CreateEvent_InvalidDate(t *testing.T) {
	h := handlers.NewEventHandler(&MockEventService{}, nil)

	req := dto.CreateEventRequest{
		Name:           "Test Event",
		Description:    "desc",
		Date:           "invalid-date",
		MaxCountPeople: 10,
		BookingTTL:     5,
		Price:          100,
	}

	w := performRequest(h.CreateEvent, "POST", "/events", req, "user-123")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestEventHandler_CreateEvent_ServiceError(t *testing.T) {
	mockEvent := &MockEventService{
		CreateEventFn: func(userid, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error) {
			return nil, errors.New("service error")
		},
	}
	h := handlers.NewEventHandler(mockEvent, nil)

	req := dto.CreateEventRequest{
		Name:           "Test Event",
		Description:    "desc",
		Date:           time.Now().Add(time.Hour).Format(time.RFC3339),
		MaxCountPeople: 10,
		BookingTTL:     5,
		Price:          100,
	}

	w := performRequest(h.CreateEvent, "POST", "/events", req, "user-123")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestEventHandler_CreateBooking_NoUser(t *testing.T) {
	h := handlers.NewEventHandler(nil, &MockBookingService{})

	req := dto.CreateBookingRequest{
		EventID:              uuid.New().String(),
		Count:                1,
		TelegramNotification: true,
		EmailNotification:    false,
	}

	w := performRequest(h.CreateBooking, "POST", "/book", req, "")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestEventHandler_CreateBooking_ServiceError(t *testing.T) {
	mockBooking := &MockBookingService{
		CreateBookingFn: func(eventId, userId string, tg, mail bool, count int) (*booking.Booking, error) {
			return nil, errors.New("service error")
		},
	}
	h := handlers.NewEventHandler(nil, mockBooking)

	req := dto.CreateBookingRequest{
		EventID:              uuid.New().String(),
		Count:                1,
		TelegramNotification: true,
		EmailNotification:    false,
	}

	w := performRequest(h.CreateBooking, "POST", "/book", req, "user123")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestEventHandler_GetEvent_WithBookings(t *testing.T) {
	b := &booking.Booking{
		ID:        uuid.New(),
		EventID:   uuid.New(),
		UserID:    uuid.New(),
		Status:    booking.BookingStatusCreated,
		Count:     2,
		Price:     100,
		ExpiredAt: time.Now().Add(10 * time.Minute),
	}
	ev := &event.Event{
		Id:             uuid.New(),
		Name:           "Test Event",
		FreePlaces:     5,
		Date:           time.Now(),
		Bookings:       []*booking.Booking{b},
		BookingTTL:     10,
		MaxCountPeople: 50,
	}

	mockEvent := &MockEventService{
		GetEventFn: func(eventid string) (*event.Event, error) { return ev, nil },
	}

	h := handlers.NewEventHandler(mockEvent, nil)

	w := performRequest(h.GetEvent, "GET", "/events/"+ev.Id.String(), nil, "u1")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
