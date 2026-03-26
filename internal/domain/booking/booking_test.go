package booking_test

import (
	"eventbooker/internal/domain/booking"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNew_Success(t *testing.T) {
	eventID := uuid.New().String()
	userID := uuid.New().String()

	b, err := booking.New(
		eventID, userID,
		"12345", "mail@mail.com",
		"Test Event",
		true, true,
		2, 30, 100.50,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b == nil {
		t.Fatal("booking is nil")
	}
	if b.EventID.String() != eventID {
		t.Error("wrong EventID")
	}
	if b.UserID.String() != userID {
		t.Error("wrong UserID")
	}
	if b.Price != 100.50*2 {
		t.Error("price not multiplied by count")
	}
	if b.Status != booking.StatusCreated {
		t.Error("wrong initial status")
	}
	if b.EventName != "Test Event" {
		t.Error("wrong event name")
	}
	if b.TelegramRecepient != "12345" {
		t.Error("wrong telegram recipient")
	}
	if b.EmailRecepient != "mail@mail.com" {
		t.Error("wrong email recipient")
	}
	if b.Count != 2 {
		t.Error("wrong count")
	}

	diff := b.ExpiredAt.Sub(b.CreatedAt)
	if diff < 29*time.Minute || diff > 31*time.Minute {
		t.Errorf("expiredAt incorrectly calculated, diff = %v", diff)
	}
}

func TestNew_InvalidEventID(t *testing.T) {
	userID := uuid.New().String()
	_, err := booking.New("invalid-uuid", userID, "12345", "mail@mail.com", "Event", false, false, 1, 10, 10)
	if err == nil {
		t.Fatal("expected error for invalid eventId")
	}
}

func TestNew_InvalidUserID(t *testing.T) {
	eventID := uuid.New().String()
	_, err := booking.New(eventID, "invalid-user-id", "12345", "mail@mail.com", "Event", false, false, 1, 10, 10)
	if err == nil {
		t.Fatal("expected error for invalid userId")
	}
}

func TestNew_InvalidCount(t *testing.T) {
	eventID := uuid.New().String()
	userID := uuid.New().String()
	_, err := booking.New(eventID, userID, "12345", "mail@mail.com", "Event", false, false, 0, 10, 10)
	if err == nil {
		t.Fatal("expected error for count <= 0")
	}
}

func TestBooking_Confirm(t *testing.T) {
	eventID := uuid.New().String()
	userID := uuid.New().String()
	b, _ := booking.New(eventID, userID, "12345", "mail@mail.com", "Event", false, false, 1, 10, 10)
	b.Confirm()
	if b.Status != booking.StatusConfirmed {
		t.Fatal("status should be confirmed")
	}
}
