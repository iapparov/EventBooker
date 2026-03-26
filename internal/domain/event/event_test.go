package event_test

import (
	"eventbooker/internal/domain/event"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNew_Success(t *testing.T) {
	creatorID := uuid.New().String()
	date := time.Now().Add(24 * time.Hour)

	e, err := event.New(creatorID, "Test Event", "Some description", date, 30, 100, 150.75)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e == nil {
		t.Fatal("event is nil")
	}
	if e.CreatorID.String() != creatorID {
		t.Error("wrong CreatorID")
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

func TestNew_InvalidCreatorID(t *testing.T) {
	date := time.Now().Add(24 * time.Hour)
	_, err := event.New("invalid-uuid", "Test", "Desc", date, 10, 10, 10)
	if err == nil {
		t.Fatal("expected error for invalid Creator ID")
	}
}

func TestNew_InvalidMaxCountPeople(t *testing.T) {
	creator := uuid.New().String()
	date := time.Now().Add(24 * time.Hour)
	_, err := event.New(creator, "Test", "Desc", date, 10, 0, 10)
	if err == nil {
		t.Fatal("expected error for MaxCountPeople <= 0")
	}
}

func TestNew_InvalidPrice(t *testing.T) {
	creator := uuid.New().String()
	date := time.Now().Add(24 * time.Hour)
	_, err := event.New(creator, "Test", "Desc", date, 10, 10, -1)
	if err == nil {
		t.Fatal("expected error for negative price")
	}
}
