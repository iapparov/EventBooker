package booking

import (
	"errors"
	"eventbooker/internal/domain/booking"
	"eventbooker/internal/domain/event"
	"eventbooker/internal/domain/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockBookingRepo struct {
	mock.Mock
}

func (m *MockBookingRepo) CreateBooking(b *booking.Booking) error {
	args := m.Called(b)
	return args.Error(0)
}

func (m *MockBookingRepo) ConfirmBooking(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockBookingRepo) GetEvent(eventId string) (*event.Event, error) {
	args := m.Called(eventId)
	return args.Get(0).(*event.Event), args.Error(1)
}

func (m *MockBookingRepo) GetUserbyUUID(id string) (*user.User, error) {
	args := m.Called(id)
	return args.Get(0).(*user.User), args.Error(1)
}

type MockBroker struct {
	mock.Mock
}

func (m *MockBroker) PublishMsg(b *booking.Booking) error {
	args := m.Called(b)
	return args.Error(0)
}

// SUCCESS CASE
func TestBookingService_CreateBooking_Success(t *testing.T) {
	repo := new(MockBookingRepo)
	broker := new(MockBroker)
	service := NewBookingService(repo, broker)

	eventID := uuid.New()
	userID := uuid.New()

	eventObj := &event.Event{
		Id:         eventID,
		Name:       "Test Event",
		FreePlaces: 10,
		Price:      100,
		BookingTTL: 1,
	}

	userObj := &user.User{
		Id:       userID,
		Telegram: "@testuser",
		Email:    "test@example.com",
	}

	repo.On("GetEvent", eventID.String()).Return(eventObj, nil)
	repo.On("GetUserbyUUID", userID.String()).Return(userObj, nil)
	repo.On("CreateBooking", mock.Anything).Return(nil)
	broker.On("PublishMsg", mock.Anything).Return(nil)

	b, err := service.CreateBooking(eventID.String(), userID.String(), true, true, 2)

	assert.NoError(t, err)
	assert.NotNil(t, b)

	repo.AssertExpectations(t)
	broker.AssertExpectations(t)
}

//
// ERROR CASES
//

func TestBookingService_CreateBooking_InvalidEventID(t *testing.T) {
	repo := new(MockBookingRepo)
	broker := new(MockBroker)
	service := NewBookingService(repo, broker)

	b, err := service.CreateBooking("invalid-id", uuid.New().String(), true, true, 1)

	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_CreateBooking_GetEventError(t *testing.T) {
	repo := new(MockBookingRepo)
	broker := new(MockBroker)
	service := NewBookingService(repo, broker)

	eventID := uuid.New().String()

	repo.On("GetEvent", eventID).Return((*event.Event)(nil), errors.New("db error"))

	b, err := service.CreateBooking(eventID, uuid.New().String(), true, true, 1)

	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_CreateBooking_InvalidUserID(t *testing.T) {
	repo := new(MockBookingRepo)
	broker := new(MockBroker)
	service := NewBookingService(repo, broker)

	eventID := uuid.New()
	invalidUserID := "invalid-uuid"

	eventObj := &event.Event{
		Id:         eventID,
		Name:       "Test Event",
		FreePlaces: 10,
		Price:      100,
		BookingTTL: 1,
	}

	repo.On("GetEvent", eventID.String()).Return(eventObj, nil)

	_, err := service.CreateBooking(eventID.String(), invalidUserID, false, false, 1)

	assert.Error(t, err)
	repo.AssertExpectations(t)
	broker.AssertExpectations(t)
}

func TestBookingService_CreateBooking_GetUserError(t *testing.T) {
	repo := new(MockBookingRepo)
	broker := new(MockBroker)
	service := NewBookingService(repo, broker)

	eventID := uuid.New().String()
	userID := uuid.New().String()

	eventObj := &event.Event{
		Id:         uuid.New(),
		FreePlaces: 5,
		Name:       "Test",
	}

	repo.On("GetEvent", eventID).Return(eventObj, nil)
	repo.On("GetUserbyUUID", userID).Return((*user.User)(nil), errors.New("user not found"))

	b, err := service.CreateBooking(eventID, userID, true, true, 1)

	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_CreateBooking_UserNil(t *testing.T) {
	repo := new(MockBookingRepo)
	broker := new(MockBroker)
	service := NewBookingService(repo, broker)

	eventID := uuid.New().String()
	userID := uuid.New().String()

	eventObj := &event.Event{
		Id:         uuid.New(),
		FreePlaces: 5,
		Name:       "Test",
	}

	repo.On("GetEvent", eventID).Return(eventObj, nil)
	repo.On("GetUserbyUUID", userID).Return((*user.User)(nil), nil)

	b, err := service.CreateBooking(eventID, userID, true, true, 1)

	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_CreateBooking_CreateBookingError(t *testing.T) {
	repo := new(MockBookingRepo)
	broker := new(MockBroker)
	service := NewBookingService(repo, broker)

	eventID := uuid.New().String()
	userID := uuid.New().String()

	eventObj := &event.Event{
		Id:         uuid.New(),
		FreePlaces: 10,
		Price:      10,
		Name:       "Test",
	}

	userObj := &user.User{
		Id:       uuid.New(),
		Telegram: "@test",
		Email:    "mail@example.com",
	}

	repo.On("GetEvent", eventID).Return(eventObj, nil)
	repo.On("GetUserbyUUID", userID).Return(userObj, nil)
	repo.On("CreateBooking", mock.Anything).Return(errors.New("insert error"))

	b, err := service.CreateBooking(eventID, userID, true, true, 2)

	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_CreateBooking_PublishMsgError(t *testing.T) {
	repo := new(MockBookingRepo)
	broker := new(MockBroker)
	service := NewBookingService(repo, broker)

	eventID := uuid.New().String()
	userID := uuid.New().String()

	eventObj := &event.Event{
		Id:         uuid.New(),
		FreePlaces: 10,
		Price:      10, // важное условие: цена > 0 → PublishMsg вызывается
		Name:       "Test",
	}

	userObj := &user.User{
		Id:       uuid.New(),
		Telegram: "@test",
		Email:    "mail@example.com",
	}

	repo.On("GetEvent", eventID).Return(eventObj, nil)
	repo.On("GetUserbyUUID", userID).Return(userObj, nil)
	repo.On("CreateBooking", mock.Anything).Return(nil)
	broker.On("PublishMsg", mock.Anything).Return(errors.New("kafka error"))

	b, err := service.CreateBooking(eventID, userID, true, true, 1)

	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestBookingService_CreateBooking_FreeEvent_NoPublish(t *testing.T) {
	repo := new(MockBookingRepo)
	broker := new(MockBroker)
	service := NewBookingService(repo, broker)

	eventID := uuid.New().String()
	userID := uuid.New().String()

	eventObj := &event.Event{
		Id:         uuid.New(),
		FreePlaces: 10,
		Price:      0, // бесплатное
		Name:       "Free Event",
	}

	userObj := &user.User{
		Id:       uuid.New(),
		Telegram: "@x",
		Email:    "y",
	}

	repo.On("GetEvent", eventID).Return(eventObj, nil)
	repo.On("GetUserbyUUID", userID).Return(userObj, nil)
	repo.On("CreateBooking", mock.Anything).Return(nil)

	b, err := service.CreateBooking(eventID, userID, true, true, 1)

	assert.NoError(t, err)
	assert.NotNil(t, b)

	// PublishMsg не вызывается
	broker.AssertNotCalled(t, "PublishMsg", mock.Anything)
}

//
// ConfirmBooking error cases
//

func TestBookingService_ConfirmBooking_ErrorRepo(t *testing.T) {
	repo := new(MockBookingRepo)
	broker := new(MockBroker)
	service := NewBookingService(repo, broker)

	id := uuid.New().String()
	repo.On("ConfirmBooking", id).Return(errors.New("db error"))

	err := service.ConfirmBooking(id)

	assert.Error(t, err)
}
