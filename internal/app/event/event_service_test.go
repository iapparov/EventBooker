package event

import (
	"errors"
	"eventbooker/internal/config"
	"eventbooker/internal/domain/event"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

// ---- MOCKS ----

type MockEventRepo struct {
	mock.Mock
}

func (m *MockEventRepo) CreateEvent(e *event.Event) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockEventRepo) GetEvent(id string) (*event.Event, error) {
	args := m.Called(id)
	return args.Get(0).(*event.Event), args.Error(1)
}

// ---- TESTS ----

func defaultCfg() *config.EventConfig {
	return &config.EventConfig{
		NameMinLength:        3,
		NameMaxLegth:         20,
		DescriptionRequire:   true,
		DesctiptionMaXLength: 100,
	}
}

//
// ------------------------ CreateEvent ------------------------
//

func TestEventService_CreateEvent_Success(t *testing.T) {
	repo := new(MockEventRepo)
	cfg := defaultCfg()
	service := NewEventService(repo, cfg)

	userID := uuid.New().String()
	date := time.Now().Add(24 * time.Hour)

	// repo expectation
	repo.On("CreateEvent", mock.Anything).Return(nil)

	e, err := service.CreateEvent(
		userID,
		"Valid Name",
		"Valid description",
		date,
		60,
		100,
		10,
	)

	assert.NoError(t, err)
	assert.NotNil(t, e)

	repo.AssertExpectations(t)
}

func TestEventService_CreateEvent_NameInvalid(t *testing.T) {
	repo := new(MockEventRepo)
	cfg := defaultCfg()
	service := NewEventService(repo, cfg)

	_, err := service.CreateEvent(
		uuid.New().String(),
		"x", // too short
		"desc",
		time.Now().Add(24*time.Hour),
		10, 10, 10,
	)

	assert.Error(t, err)
}

func TestEventService_CreateEvent_DescriptionRequired(t *testing.T) {
	repo := new(MockEventRepo)
	cfg := defaultCfg()
	service := NewEventService(repo, cfg)

	_, err := service.CreateEvent(
		uuid.New().String(),
		"Valid Name",
		"", // required but empty
		time.Now().Add(24*time.Hour),
		10, 10, 10,
	)

	assert.Error(t, err)
}

func TestEventService_CreateEvent_DescriptionTooLong(t *testing.T) {
	repo := new(MockEventRepo)
	cfg := defaultCfg()
	service := NewEventService(repo, cfg)

	longDescr := ""
	for i := 0; i < 200; i++ {
		longDescr += "a"
	}

	_, err := service.CreateEvent(
		uuid.New().String(),
		"Valid Name",
		longDescr,
		time.Now().Add(24*time.Hour),
		10, 10, 10,
	)

	assert.Error(t, err)
}

func TestEventService_CreateEvent_DateInPast(t *testing.T) {
	repo := new(MockEventRepo)
	cfg := defaultCfg()
	service := NewEventService(repo, cfg)

	_, err := service.CreateEvent(
		uuid.New().String(),
		"Valid Name",
		"Valid desc",
		time.Now().Add(-1*time.Hour),
		10, 10, 10,
	)

	assert.Error(t, err)
}

func TestEventService_CreateEvent_FreeEventTTLForcedZero(t *testing.T) {
	repo := new(MockEventRepo)
	cfg := defaultCfg()
	service := NewEventService(repo, cfg)

	repo.On("CreateEvent", mock.Anything).Return(nil)

	e, err := service.CreateEvent(
		uuid.New().String(),
		"Valid Name",
		"Valid desc",
		time.Now().Add(24*time.Hour),
		99, // will be ignored
		10,
		0, // free
	)

	assert.NoError(t, err)
	assert.Equal(t, 0, e.BookingTTL)

	repo.AssertExpectations(t)
}

func TestEventService_CreateEvent_InvalidTTL(t *testing.T) {
	repo := new(MockEventRepo)
	cfg := defaultCfg()
	service := NewEventService(repo, cfg)

	_, err := service.CreateEvent(
		uuid.New().String(),
		"Valid Name",
		"Valid desc",
		time.Now().Add(24*time.Hour),
		-5, // invalid
		10,
		10,
	)

	assert.Error(t, err)
}

func TestEventService_CreateEvent_RepoError(t *testing.T) {
	repo := new(MockEventRepo)
	cfg := defaultCfg()
	service := NewEventService(repo, cfg)

	repo.On("CreateEvent", mock.Anything).Return(errors.New("db_error"))

	_, err := service.CreateEvent(
		uuid.New().String(),
		"Valid Name",
		"Valid desc",
		time.Now().Add(24*time.Hour),
		10, 10, 10,
	)

	assert.Error(t, err)
	repo.AssertExpectations(t)
}

//
// ------------------------ GetEvent ------------------------
//

func TestEventService_GetEvent_Success(t *testing.T) {
	repo := new(MockEventRepo)
	cfg := defaultCfg()
	service := NewEventService(repo, cfg)

	eventID := uuid.New().String()
	ev := &event.Event{Id: uuid.MustParse(eventID)}

	repo.On("GetEvent", eventID).Return(ev, nil)

	result, err := service.GetEvent(eventID)

	assert.NoError(t, err)
	assert.Equal(t, ev, result)

	repo.AssertExpectations(t)
}

func TestEventService_GetEvent_InvalidUUID(t *testing.T) {
	repo := new(MockEventRepo)
	cfg := defaultCfg()
	service := NewEventService(repo, cfg)

	_, err := service.GetEvent("invalid-uuid")
	assert.Error(t, err)
}

func TestEventService_GetEvent_RepoError(t *testing.T) {
	repo := new(MockEventRepo)
	cfg := defaultCfg()
	service := NewEventService(repo, cfg)

	id := uuid.New().String()

	repo.On("GetEvent", id).Return(&event.Event{}, errors.New("db error"))

	_, err := service.GetEvent(id)
	assert.Error(t, err)
}

//
// ------------------------ isNameValid ------------------------
//

func TestEventService_isNameValid(t *testing.T) {
	cfg := defaultCfg()
	s := NewEventService(nil, cfg)

	assert.Error(t, s.isNameValid(""))                      // empty
	assert.Error(t, s.isNameValid("ab"))                    // too short
	assert.Error(t, s.isNameValid("aaaaaaaaaaaaaaaaaaaaa")) // too long
	assert.NoError(t, s.isNameValid("Valid"))
}

//
// ------------------------ isDescriptionValid ------------------------
//

func TestEventService_isDescriptionValid(t *testing.T) {
	cfg := defaultCfg()
	s := NewEventService(nil, cfg)

	assert.Error(t, s.isDescriptionValid("")) // required

	long := ""
	for i := 0; i < 200; i++ {
		long += "a"
	}

	assert.Error(t, s.isDescriptionValid(long))
	assert.NoError(t, s.isDescriptionValid("valid description"))
}
