package event_test

import (
	event "eventbooker/internal/domain/event"
	"github.com/google/uuid"
	"testing"
	"time"
)

//
// ------------------ SUCCESS ------------------
//

func TestNewEvent_Success(t *testing.T) {
	creatorID := uuid.New().String()
	date := time.Now().Add(24 * time.Hour)

	e, err := event.NewEvent(
		creatorID,
		"Test Event",
		"Some description",
		date,
		30,     // bookingTTL
		100,    // MaxCountPeople
		150.75, // Price
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e == nil {
		t.Fatal("event is nil")
	}

	if e.CreatorId.String() != creatorID {
		t.Error("wrong CreatorId")
	}
	if e.Name != "Test Event" {
		t.Error("wrong name")
	}
	if e.Description != "Some description" {
		t.Error("wrong description")
	}
	if e.MaxCountPeople != 100 {
		t.Error("wrong MaxCountPeople")
	}
	if e.FreePlaces != 100 {
		t.Error("FreePlaces must equal MaxCountPeople")
	}
	if e.Price != 150.75 {
		t.Error("wrong price")
	}
	if e.BookingTTL != 30 {
		t.Error("wrong BookingTTL")
	}
	if e.Date != date {
		t.Error("wrong date")
	}
	if len(e.Bookings) != 0 {
		t.Error("Bookings must be initialized as an empty slice")
	}
}

//
// ------------------ ERROR: INVALID CREATOR UUID ------------------
//

func TestNewEvent_InvalidCreatorID(t *testing.T) {
	date := time.Now().Add(24 * time.Hour)

	_, err := event.NewEvent(
		"invalid-uuid",
		"Test",
		"Desc",
		date,
		10,
		10,
		10,
	)

	if err == nil {
		t.Fatal("expected error for invalid Creator ID")
	}
}

//
// ------------------ ERROR: MaxCountPeople <= 0 ------------------
//

func TestNewEvent_InvalidMaxCountPeople(t *testing.T) {
	creator := uuid.New().String()
	date := time.Now().Add(24 * time.Hour)

	_, err := event.NewEvent(
		creator,
		"Test",
		"Desc",
		date,
		10,
		0, // invalid
		10,
	)

	if err == nil {
		t.Fatal("expected error for MaxCountPeople <= 0")
	}
}

//
// ------------------ ERROR: Price < 0 ------------------
//

func TestNewEvent_InvalidPrice(t *testing.T) {
	creator := uuid.New().String()
	date := time.Now().Add(24 * time.Hour)

	_, err := event.NewEvent(
		creator,
		"Test",
		"Desc",
		date,
		10,
		10,
		-1, // invalid price
	)

	if err == nil {
		t.Fatal("expected error for negative price")
	}
}
