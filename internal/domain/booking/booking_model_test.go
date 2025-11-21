package booking_test

import (
	booking "eventbooker/internal/domain/booking"
	"github.com/google/uuid"
	"testing"
	"time"
)

//
// ------------------ SUCCESS ------------------
//

func TestNewBooking_Success(t *testing.T) {
	eventID := uuid.New().String()
	userID := uuid.New().String()

	b, err := booking.NewBooking(
		eventID,
		userID,
		"12345", "mail@mail.com",
		"Test Event",
		true, true,
		2,      // count
		30,     // expiredAt min
		100.50, // price
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
	if b.Status != booking.BookingStatusCreated {
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

	// expiredAt ~ now + 30 minutes
	diff := b.ExpiredAt.Sub(b.CreatedAt)
	if diff < 29*time.Minute || diff > 31*time.Minute {
		t.Errorf("expiredAt incorrectly calculated, diff = %v", diff)
	}
}

//
// ------------------ ERROR: BAD EVENT ID ------------------
//

func TestNewBooking_InvalidEventID(t *testing.T) {
	userID := uuid.New().String()

	_, err := booking.NewBooking(
		"invalid-uuid",
		userID,
		"12345", "mail@mail.com",
		"Event",
		false, false,
		1, 10, 10,
	)

	if err == nil {
		t.Fatal("expected error for invalid eventId")
	}
}

//
// ------------------ ERROR: BAD USER ID ------------------
//

func TestNewBooking_InvalidUserID(t *testing.T) {
	eventID := uuid.New().String()

	_, err := booking.NewBooking(
		eventID,
		"invalid-user-id",
		"12345", "mail@mail.com",
		"Event",
		false, false,
		1, 10, 10,
	)

	if err == nil {
		t.Fatal("expected error for invalid userId")
	}
}

//
// ------------------ ERROR: COUNT <= 0 ------------------
//

func TestNewBooking_InvalidCount(t *testing.T) {
	eventID := uuid.New().String()
	userID := uuid.New().String()

	_, err := booking.NewBooking(
		eventID,
		userID,
		"12345", "mail@mail.com",
		"Event",
		false, false,
		0, // invalid count
		10, 10,
	)

	if err == nil {
		t.Fatal("expected error for count <= 0")
	}
}

//
// ------------------ STATUS CONFIRMED ------------------
//

func TestBooking_StatusConfirmed(t *testing.T) {
	eventID := uuid.New().String()
	userID := uuid.New().String()

	b, _ := booking.NewBooking(
		eventID,
		userID,
		"12345", "mail@mail.com",
		"Event",
		false, false,
		1, 10, 10,
	)

	b.StatusConfirmed()

	if b.Status != booking.BookingStatusConfirmed {
		t.Fatal("status should be confirmed")
	}
}
