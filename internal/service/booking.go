package service

import (
	"context"
	"errors"
	"fmt"

	"eventbooker/internal/domain/booking"
	"eventbooker/internal/domain/event"
	"eventbooker/internal/domain/user"

	"github.com/google/uuid"
	wbzlog "github.com/wb-go/wbf/zlog"
)

// BookingRepository defines the storage operations needed by BookingService.
type BookingRepository interface {
	CreateBooking(ctx context.Context, b *booking.Booking) error
	ConfirmBooking(ctx context.Context, id string) error
	GetEvent(ctx context.Context, eventID string) (*event.Event, error)
	GetUserByUUID(ctx context.Context, id string) (*user.User, error)
}

// BookingBroker defines the message broker operations needed by BookingService.
type BookingBroker interface {
	PublishMsg(ctx context.Context, b *booking.Booking) error
}

// BookingService handles booking business logic.
type BookingService struct {
	repo   BookingRepository
	broker BookingBroker
}

// NewBookingService creates a new BookingService.
func NewBookingService(repo BookingRepository, broker BookingBroker) *BookingService {
	return &BookingService{
		repo:   repo,
		broker: broker,
	}
}

// Create creates a new booking.
func (s *BookingService) Create(ctx context.Context, eventID, userID string, telegramNotification, emailNotification bool, count int) (*booking.Booking, error) {
	if _, err := uuid.Parse(eventID); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid event id")
		return nil, err
	}

	ev, err := s.repo.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if ev.FreePlaces-count < 0 {
		return nil, errors.New("not enough available seats")
	}

	if _, err := uuid.Parse(userID); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid user id")
		return nil, err
	}

	u, err := s.repo.GetUserByUUID(ctx, userID)
	if err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("user not found")
		return nil, fmt.Errorf("user not found")
	}
	if u == nil {
		wbzlog.Logger.Debug().Msg("user is nil")
		return nil, fmt.Errorf("user not found")
	}

	b, err := booking.New(eventID, userID, u.Telegram, u.Email, ev.Name, telegramNotification, emailNotification, count, ev.BookingTTL, ev.Price)
	if err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("cannot create booking")
		return nil, err
	}

	if ev.Price == 0 {
		b.Confirm()
	}

	if err = s.repo.CreateBooking(ctx, b); err != nil {
		return nil, err
	}

	if ev.Price != 0 {
		if err = s.broker.PublishMsg(ctx, b); err != nil {
			return nil, err
		}
	}

	return b, nil
}

// Confirm confirms an existing booking.
func (s *BookingService) Confirm(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid booking id")
		return err
	}

	return s.repo.ConfirmBooking(ctx, id)
}
