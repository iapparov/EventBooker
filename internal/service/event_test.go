package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"eventbooker/internal/config"
	"eventbooker/internal/domain/event"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEventRepo struct{ mock.Mock }

func (m *mockEventRepo) CreateEvent(ctx context.Context, e *event.Event) error {
	return m.Called(e).Error(0)
}
func (m *mockEventRepo) GetEvent(ctx context.Context, id string) (*event.Event, error) {
	args := m.Called(id)
	return args.Get(0).(*event.Event), args.Error(1)
}

func defaultEventCfg() *config.EventConfig {
	return &config.EventConfig{
		NameMinLength:        3,
		NameMaxLength:        20,
		DescriptionRequired:  true,
		DescriptionMaxLength: 100,
	}
}

func TestEventService_Create_Success(t *testing.T) {
	repo := new(mockEventRepo)
	cfg := defaultEventCfg()
	svc := NewEventService(repo, cfg)
	repo.On("CreateEvent", mock.Anything).Return(nil)
	e, err := svc.Create(context.Background(), uuid.New().String(), "Valid Name", "Valid description", time.Now().Add(24*time.Hour), 60, 100, 10)
	assert.NoError(t, err)
	assert.NotNil(t, e)
	repo.AssertExpectations(t)
}

func TestEventService_Create_NameInvalid(t *testing.T) {
	svc := NewEventService(new(mockEventRepo), defaultEventCfg())
	_, err := svc.Create(context.Background(), uuid.New().String(), "x", "desc", time.Now().Add(24*time.Hour), 10, 10, 10)
	assert.Error(t, err)
}

func TestEventService_Create_DescriptionRequired(t *testing.T) {
	svc := NewEventService(new(mockEventRepo), defaultEventCfg())
	_, err := svc.Create(context.Background(), uuid.New().String(), "Valid Name", "", time.Now().Add(24*time.Hour), 10, 10, 10)
	assert.Error(t, err)
}

func TestEventService_Create_DescriptionTooLong(t *testing.T) {
	svc := NewEventService(new(mockEventRepo), defaultEventCfg())
	longDescr := ""
	for i := 0; i < 200; i++ {
		longDescr += "a"
	}
	_, err := svc.Create(context.Background(), uuid.New().String(), "Valid Name", longDescr, time.Now().Add(24*time.Hour), 10, 10, 10)
	assert.Error(t, err)
}

func TestEventService_Create_DateInPast(t *testing.T) {
	svc := NewEventService(new(mockEventRepo), defaultEventCfg())
	_, err := svc.Create(context.Background(), uuid.New().String(), "Valid Name", "Valid desc", time.Now().Add(-1*time.Hour), 10, 10, 10)
	assert.Error(t, err)
}

func TestEventService_Create_FreeEventTTLForcedZero(t *testing.T) {
	repo := new(mockEventRepo)
	svc := NewEventService(repo, defaultEventCfg())
	repo.On("CreateEvent", mock.Anything).Return(nil)
	e, err := svc.Create(context.Background(), uuid.New().String(), "Valid Name", "Valid desc", time.Now().Add(24*time.Hour), 99, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, e.BookingTTL)
	repo.AssertExpectations(t)
}

func TestEventService_Create_InvalidTTL(t *testing.T) {
	svc := NewEventService(new(mockEventRepo), defaultEventCfg())
	_, err := svc.Create(context.Background(), uuid.New().String(), "Valid Name", "Valid desc", time.Now().Add(24*time.Hour), -5, 10, 10)
	assert.Error(t, err)
}

func TestEventService_Create_RepoError(t *testing.T) {
	repo := new(mockEventRepo)
	svc := NewEventService(repo, defaultEventCfg())
	repo.On("CreateEvent", mock.Anything).Return(errors.New("db_error"))
	_, err := svc.Create(context.Background(), uuid.New().String(), "Valid Name", "Valid desc", time.Now().Add(24*time.Hour), 10, 10, 10)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestEventService_Get_Success(t *testing.T) {
	repo := new(mockEventRepo)
	svc := NewEventService(repo, defaultEventCfg())
	eventID := uuid.New().String()
	ev := &event.Event{ID: uuid.MustParse(eventID)}
	repo.On("GetEvent", eventID).Return(ev, nil)
	result, err := svc.Get(context.Background(), eventID)
	assert.NoError(t, err)
	assert.Equal(t, ev, result)
	repo.AssertExpectations(t)
}

func TestEventService_Get_InvalidUUID(t *testing.T) {
	svc := NewEventService(new(mockEventRepo), defaultEventCfg())
	_, err := svc.Get(context.Background(), "invalid-uuid")
	assert.Error(t, err)
}

func TestEventService_Get_RepoError(t *testing.T) {
	repo := new(mockEventRepo)
	svc := NewEventService(repo, defaultEventCfg())
	id := uuid.New().String()
	repo.On("GetEvent", id).Return(&event.Event{}, errors.New("db error"))
	_, err := svc.Get(context.Background(), id)
	assert.Error(t, err)
}

func TestEventService_validateName(t *testing.T) {
	svc := NewEventService(nil, defaultEventCfg())
	assert.Error(t, svc.validateName(""))
	assert.Error(t, svc.validateName("ab"))
	assert.Error(t, svc.validateName("aaaaaaaaaaaaaaaaaaaaa"))
	assert.NoError(t, svc.validateName("Valid"))
}

func TestEventService_validateDescription(t *testing.T) {
	svc := NewEventService(nil, defaultEventCfg())
	assert.Error(t, svc.validateDescription(""))
	long := ""
	for i := 0; i < 200; i++ {
		long += "a"
	}
	assert.Error(t, svc.validateDescription(long))
	assert.NoError(t, svc.validateDescription("valid description"))
}
