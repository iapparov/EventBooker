package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"eventbooker/internal/config"
	"eventbooker/internal/domain/event"

	"github.com/google/uuid"
	wbzlog "github.com/wb-go/wbf/zlog"
)

// EventRepository defines the storage operations needed by EventService.
type EventRepository interface {
	CreateEvent(ctx context.Context, e *event.Event) error
	GetEvent(ctx context.Context, eventID string) (*event.Event, error)
}

// EventService handles event business logic.
type EventService struct {
	repo EventRepository
	cfg  *config.EventConfig
}

// NewEventService creates a new EventService.
func NewEventService(repo EventRepository, cfg *config.EventConfig) *EventService {
	return &EventService{
		repo: repo,
		cfg:  cfg,
	}
}

// Create creates a new event.
func (s *EventService) Create(ctx context.Context, userID, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error) {
	if err := s.validateName(name); err != nil {
		wbzlog.Logger.Debug().Err(err)
		return nil, err
	}

	if err := s.validateDescription(description); err != nil {
		wbzlog.Logger.Debug().Err(err)
		return nil, err
	}

	if date.Before(time.Now()) {
		wbzlog.Logger.Debug().Msg("date should be in the future")
		return nil, fmt.Errorf("event date must be in the future")
	}

	if price == 0 {
		bookingTTL = 0
	}

	if bookingTTL < 0 {
		return nil, errors.New("booking TTL must be >= 0")
	}

	ev, err := event.New(userID, name, description, date, bookingTTL, maxCountPeople, price)
	if err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("create event error (domain level)")
		return nil, err
	}

	if err = s.repo.CreateEvent(ctx, ev); err != nil {
		return nil, err
	}

	return ev, nil
}

// Get retrieves an event by ID.
func (s *EventService) Get(ctx context.Context, eventID string) (*event.Event, error) {
	if _, err := uuid.Parse(eventID); err != nil {
		wbzlog.Logger.Debug().Err(err).Msgf("invalid event_id: %s", eventID)
		return nil, err
	}

	return s.repo.GetEvent(ctx, eventID)
}

func (s *EventService) validateName(name string) error {
	l := utf8.RuneCountInString(name)
	if name == "" || l < s.cfg.NameMinLength || l > s.cfg.NameMaxLength {
		return fmt.Errorf("name must be between %d and %d characters", s.cfg.NameMinLength, s.cfg.NameMaxLength)
	}
	return nil
}

func (s *EventService) validateDescription(descr string) error {
	if s.cfg.DescriptionRequired && descr == "" {
		return errors.New("description required")
	}
	if utf8.RuneCountInString(descr) > s.cfg.DescriptionMaxLength {
		return fmt.Errorf("description must be shorter than %d characters", s.cfg.DescriptionMaxLength)
	}
	return nil
}
