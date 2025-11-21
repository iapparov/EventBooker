package event

import (
	"errors"
	"eventbooker/internal/config"
	"eventbooker/internal/domain/event"
	"fmt"
	"time"
	"unicode/utf8"
	"github.com/google/uuid"
	wbzlog "github.com/wb-go/wbf/zlog"
)

type EventService struct {
	repo EventStorageProvider
	cfg *config.EventConfig
}

type EventStorageProvider interface {
	CreateEvent(event *event.Event) error
	GetEvent(evetnid string) (*event.Event, error)
}

func NewEventService (repo EventStorageProvider, cfg *config.EventConfig) *EventService {
	return &EventService{
		repo: repo,
		cfg: cfg,
	}
}

func (s *EventService) CreateEvent(userid, name, description string, date time.Time, bookingTTL, maxCountPeople int, price float64) (*event.Event, error) {
	err := s.isNameValid(name)
	if err != nil {
		wbzlog.Logger.Debug().Err(err)
		return nil, err
	}

	err = s.isDescriptionValid(description)
	if err != nil {
		wbzlog.Logger.Debug().Err(err)
		return nil, err
	}

	if date.Before(time.Now()) {
		wbzlog.Logger.Debug().Msg("date should be bigger than today")
		return nil, fmt.Errorf("event date must be in the future")
	}
	if price == 0 {
		bookingTTL = 0 // Если цена бесплатно, то бронь автоподтверждается
	}
	if bookingTTL < 0 {
		return nil, errors.New("booking TTL must be bigger or equal 0")
	}
	event, err := event.NewEvent(userid, name, description, date, bookingTTL, maxCountPeople, price)
	if err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("create event error (domain_level)")
		return nil, err
	}
	err = s.repo.CreateEvent(event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) GetEvent(eventid string) (*event.Event, error) {
	_, err := uuid.Parse(eventid)
	if err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid event_id")
		return nil, err
	}
	return s.repo.GetEvent(eventid)
}

func (s* EventService) isNameValid(name string) error {
	if name == "" || utf8.RuneCountInString(name) < s.cfg.NameMinLength || utf8.RuneCountInString(name) > s.cfg.NameMaxLegth {
		return fmt.Errorf("name cant be empty, should be bigger than %d, and smaller than %d", s.cfg.NameMinLength, s.cfg.NameMaxLegth)
	}
	return nil
}

func (s* EventService) isDescriptionValid(descr string) error {
	if s.cfg.DescriptionRequire && descr == ""{
		return errors.New("description required")
	}
	if utf8.RuneCountInString(descr) > s.cfg.DesctiptionMaXLength {
		return fmt.Errorf("description should be smaller than %d", s.cfg.DesctiptionMaXLength)
	}
	return nil
}