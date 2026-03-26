package service

import (
	"context"
	"errors"
	"testing"

	"eventbooker/internal/domain/booking"
	"eventbooker/internal/domain/event"
	"eventbooker/internal/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBookingRepo struct{ mock.Mock }

func (m *mockBookingRepo) CreateBooking(ctx context.Context, b *booking.Booking) error {
	return m.Called(b).Error(0)
}
func (m *mockBookingRepo) ConfirmBooking(ctx context.Context, id string) error {
	return m.Called(id).Error(0)
}
func (m *mockBookingRepo) GetEvent(ctx context.Context, eventID string) (*event.Event, error) {
	args := m.Called(eventID)
	return args.Get(0).(*event.Event), args.Error(1)
}
func (m *mockBookingRepo) GetUserByUUID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(id)
	return args.Get(0).(*user.User), args.Error(1)
}

type mockBroker struct{ mock.Mock }

func (m *mockBroker) PublishMsg(ctx context.Context, b *booking.Booking) error {
	return m.Called(b).Error(0)
}

func TestBookingService_Create_Success(t *testing.T) {
	repo := new(mockBookingRepo)
	broker := new(mockBroker)
	svc := NewBookingService(repo, broker)

	eventID := uuid.New()
	userID := uuid.New()

	ev := &event.Event{ID: eventID, Name: "Test", FreePlaces: 10, Price: 100, BookingTTL: 1}
	u := &user.User{ID: userID, Telegram: "@test", Email: "test@example.com"}

	repo.On("GetEvent", eventID.String()).Return(ev, nil)
	repo.On("GetUserByUUID", userID.String()).Return(u, nil)
	repo.On("CreateBooking", mock.Anything).Return(nil)
	broker.On("PublishMsg", mock.Anything).Return(nil)

	b, err := svc.Create(context.Background(), eventID.String(), userID.String(), true, true, 2)
	assert.NoError(t, err)
	assert.NotNil(t, b)
	repo.AssertExpectations(t)
	broker.AssertExpectations(t)
}

func TestBookingService_Create_InvalidEventID(t *testing.T) {
	svc := NewBookingService(new(mockBookingRepo), new(mockBroker))
	b, err := svc.Create(context.Background(), "invalid-id", uuid.New().String(), true, true, 1)
	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_Create_GetEventError(t *testing.T) {
	repo := new(mockBookingRepo)
	svc := NewBookingService(repo, new(mockBroker))
	eventID := uuid.New().String()
	repo.On("GetEvent", eventID).Return((*event.Event)(nil), errors.New("db error"))
	b, err := svc.Create(context.Background(), eventID, uuid.New().String(), true, true, 1)
	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_Create_InvalidUserID(t *testing.T) {
	repo := new(mockBookingRepo)
	svc := NewBookingService(repo, new(mockBroker))
	eventID := uuid.New()
	ev := &event.Event{ID: eventID, Name: "Test", FreePlaces: 10, Price: 100, BookingTTL: 1}
	repo.On("GetEvent", eventID.String()).Return(ev, nil)
	_, err := svc.Create(context.Background(), eventID.String(), "invalid-uuid", false, false, 1)
	assert.Error(t, err)
}

func TestBookingService_Create_GetUserError(t *testing.T) {
	repo := new(mockBookingRepo)
	svc := NewBookingService(repo, new(mockBroker))
	eventID := uuid.New().String()
	userID := uuid.New().String()
	ev := &event.Event{ID: uuid.New(), FreePlaces: 5, Name: "Test"}
	repo.On("GetEvent", eventID).Return(ev, nil)
	repo.On("GetUserByUUID", userID).Return((*user.User)(nil), errors.New("user not found"))
	b, err := svc.Create(context.Background(), eventID, userID, true, true, 1)
	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_Create_UserNil(t *testing.T) {
	repo := new(mockBookingRepo)
	svc := NewBookingService(repo, new(mockBroker))
	eventID := uuid.New().String()
	userID := uuid.New().String()
	ev := &event.Event{ID: uuid.New(), FreePlaces: 5, Name: "Test"}
	repo.On("GetEvent", eventID).Return(ev, nil)
	repo.On("GetUserByUUID", userID).Return((*user.User)(nil), nil)
	b, err := svc.Create(context.Background(), eventID, userID, true, true, 1)
	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_Create_CreateBookingError(t *testing.T) {
	repo := new(mockBookingRepo)
	broker := new(mockBroker)
	svc := NewBookingService(repo, broker)
	eventID := uuid.New().String()
	userID := uuid.New().String()
	ev := &event.Event{ID: uuid.New(), FreePlaces: 10, Price: 10, Name: "Test"}
	u := &user.User{ID: uuid.New(), Telegram: "@test", Email: "mail@example.com"}
	repo.On("GetEvent", eventID).Return(ev, nil)
	repo.On("GetUserByUUID", userID).Return(u, nil)
	repo.On("CreateBooking", mock.Anything).Return(errors.New("insert error"))
	b, err := svc.Create(context.Background(), eventID, userID, true, true, 2)
	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_Create_PublishMsgError(t *testing.T) {
	repo := new(mockBookingRepo)
	broker := new(mockBroker)
	svc := NewBookingService(repo, broker)
	eventID := uuid.New().String()
	userID := uuid.New().String()
	ev := &event.Event{ID: uuid.New(), FreePlaces: 10, Price: 10, Name: "Test"}
	u := &user.User{ID: uuid.New(), Telegram: "@test", Email: "mail@example.com"}
	repo.On("GetEvent", eventID).Return(ev, nil)
	repo.On("GetUserByUUID", userID).Return(u, nil)
	repo.On("CreateBooking", mock.Anything).Return(nil)
	broker.On("PublishMsg", mock.Anything).Return(errors.New("broker error"))
	b, err := svc.Create(context.Background(), eventID, userID, true, true, 1)
	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_Create_FreeEvent_NoPublish(t *testing.T) {
	repo := new(mockBookingRepo)
	broker := new(mockBroker)
	svc := NewBookingService(repo, broker)
	eventID := uuid.New().String()
	userID := uuid.New().String()
	ev := &event.Event{ID: uuid.New(), FreePlaces: 10, Price: 0, Name: "Free Event"}
	u := &user.User{ID: uuid.New(), Telegram: "@x", Email: "y"}
	repo.On("GetEvent", eventID).Return(ev, nil)
	repo.On("GetUserByUUID", userID).Return(u, nil)
	repo.On("CreateBooking", mock.Anything).Return(nil)
	b, err := svc.Create(context.Background(), eventID, userID, true, true, 1)
	assert.NoError(t, err)
	assert.NotNil(t, b)
	broker.AssertNotCalled(t, "PublishMsg", mock.Anything)
}

func TestBookingService_Confirm_ErrorRepo(t *testing.T) {
	repo := new(mockBookingRepo)
	svc := NewBookingService(repo, new(mockBroker))
	id := uuid.New().String()
	repo.On("ConfirmBooking", id).Return(errors.New("db error"))
	err := svc.Confirm(context.Background(), id)
	assert.Error(t, err)
}
