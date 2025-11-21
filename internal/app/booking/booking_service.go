package booking

import (
	"errors"
	"eventbooker/internal/domain/booking"
	"eventbooker/internal/domain/event"
	"eventbooker/internal/domain/user"
	"github.com/google/uuid"
	wbzlog "github.com/wb-go/wbf/zlog"
)

type BookingService struct {
	bookingRepo BookingStorageProvider
	broker BookingBrokerProvider
}

type BookingStorageProvider interface {
	CreateBooking(booking *booking.Booking) error
	ConfirmBooking(id string) error
	GetEvent(evetnid string) (*event.Event, error)
	GetUser(id string) (*user.User, error)
}

type BookingBrokerProvider interface {
	PublishMsg(booking *booking.Booking) error
}

func NewBookingService (repo BookingStorageProvider, broker BookingBrokerProvider) *BookingService {
	return &BookingService{
		bookingRepo: repo,
		broker: broker,
	}
}

func (s *BookingService) CreateBooking(eventId, userId string, telegramNotification, emailNotification bool, count int) (*booking.Booking, error){
	_, err := uuid.Parse(eventId)
	if err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid event id")
		return nil, err
	}
	event, err := s.bookingRepo.GetEvent(eventId)
	if err != nil {
		return nil, err
	}
	if event.FreePlaces - count < 0 {
		err = errors.New("dont enough available seats in reservation")
		return nil, err
	}

	_, err = uuid.Parse(userId)
	if err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid user id")
		return nil, err
	}
	user, err := s.bookingRepo.GetUser(userId)
	
	booking, err := booking.NewBooking(eventId, userId,user.Telegram, user.Email, event.Name, telegramNotification, emailNotification, count, event.BookingTTL, event.Price)
	if err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("cant create booking")
		return nil, err
	}
	if event.Price == 0 {
		booking.StatusConfirmed()
	}
	err = s.bookingRepo.CreateBooking(booking)
	if err != nil {
		return nil, err
	}

	if event.Price != 0 {
		err = s.broker.PublishMsg(booking)
		if err != nil {
			return nil, err
		}
	}
	return booking, nil
}

func (s *BookingService) ConfirmBooking(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid booking id")
		return err
	}
	err = s.bookingRepo.ConfirmBooking(id)
	if err != nil {
		return err
	}
	return nil
}
