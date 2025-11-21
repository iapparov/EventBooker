package user

import (
	"errors"
	"eventbooker/internal/auth"
	"eventbooker/internal/config"
	"eventbooker/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

// -------------------- MOCKS --------------------

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetUser(login string) (*user.User, error) {
	args := m.Called(login)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepo) SaveUser(u *user.User) error {
	args := m.Called(u)
	return args.Error(0)
}

type MockJWT struct {
	mock.Mock
}

func (m *MockJWT) GenerateTokens(u *user.User) (*auth.JWTResponse, error) {
	args := m.Called(u)
	return args.Get(0).(*auth.JWTResponse), args.Error(1)
}

func (m *MockJWT) ValidateTokens(t string) (*auth.JWTPayload, error) {
	args := m.Called(t)
	return args.Get(0).(*auth.JWTPayload), args.Error(1)
}

func (m *MockJWT) RefreshTokens(r string) (*auth.JWTResponse, error) {
	args := m.Called(r)
	return args.Get(0).(*auth.JWTResponse), args.Error(1)
}

func defaultCfg() *config.AppConfig {
	return &config.AppConfig{
		UserConfig: config.UserConfig{
			MinLength:         3,
			MaxLength:         20,
			AllowedCharacters: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-",
		},
		PasswordConfig: config.PasswordConfig{
			MinLength:    6,
			MaxLength:    32,
			RequireUpper: true,
			RequireLower: true,
			RequireDigit: true,
		},
	}
}

// -------------------- LOGIN --------------------

func TestUserService_Login_Success(t *testing.T) {
	repo := new(MockUserRepo)
	jwt := new(MockJWT)
	cfg := defaultCfg()

	service := NewUserService(repo, jwt, cfg)

	hash, _ := bcrypt.GenerateFromPassword([]byte("Password1"), bcrypt.DefaultCost)

	u := &user.User{
		Login:    "test",
		Password: hash,
	}

	repo.On("GetUser", "test").Return(u, nil)
	jwtResp := &auth.JWTResponse{AccessToken: "token"}
	jwt.On("GenerateTokens", u).Return(jwtResp, nil)

	resp, err := service.Login("test", "Password1")

	assert.NoError(t, err)
	assert.Equal(t, "token", resp.AccessToken)

	repo.AssertExpectations(t)
	jwt.AssertExpectations(t)
}

func TestUserService_Login_EmptyFields(t *testing.T) {
	repo := new(MockUserRepo)
	jwt := new(MockJWT)
	cfg := defaultCfg()

	service := NewUserService(repo, jwt, cfg)

	resp, err := service.Login("", "pass")
	assert.Error(t, err)
	assert.Nil(t, resp)

	resp2, err2 := service.Login("login", "")
	assert.Error(t, err2)
	assert.Nil(t, resp2)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := new(MockUserRepo)
	jwt := new(MockJWT)
	cfg := defaultCfg()

	service := NewUserService(repo, jwt, cfg)

	repo.On("GetUser", "test").Return(&user.User{}, errors.New("user not found"))

	resp, err := service.Login("test", "pass")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUserService_Login_InvalidPassword(t *testing.T) {
	repo := new(MockUserRepo)
	jwt := new(MockJWT)
	cfg := defaultCfg()
	service := NewUserService(repo, jwt, cfg)

	hash, _ := bcrypt.GenerateFromPassword([]byte("Password1"), bcrypt.DefaultCost)
	u := &user.User{Password: hash}

	repo.On("GetUser", "test").Return(u, nil)

	resp, err := service.Login("test", "wrong")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// -------------------- REGISTRATION --------------------

func TestUserService_Registration_Success(t *testing.T) {
	repo := new(MockUserRepo)
	jwt := new(MockJWT)
	cfg := defaultCfg()
	service := NewUserService(repo, jwt, cfg)

	repo.On("GetUser", "newuser").Return((*user.User)(nil), errors.New("user not found"))
	repo.On("SaveUser", mock.Anything).Return(nil)

	u, err := service.Registration(
		"newuser",
		"Password1",
		"email@test.com",
		"12345",
	)

	assert.NoError(t, err)
	assert.Equal(t, "newuser", u.Login)

	repo.AssertExpectations(t)
}

func TestUserService_Registration_InvalidLogin(t *testing.T) {
	service := NewUserService(new(MockUserRepo), new(MockJWT), defaultCfg())

	u, err := service.Registration("ab", "Password1", "email@test.com", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Registration_InvalidPassword(t *testing.T) {
	service := NewUserService(new(MockUserRepo), new(MockJWT), defaultCfg())

	u, err := service.Registration("validUser", "short", "email@test.com", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Registration_InvalidTelegram(t *testing.T) {
	service := NewUserService(new(MockUserRepo), new(MockJWT), defaultCfg())

	u, err := service.Registration("validUser", "Password1", "email@test.com", "abc")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Registration_InvalidEmail(t *testing.T) {
	service := NewUserService(new(MockUserRepo), new(MockJWT), defaultCfg())

	u, err := service.Registration("validUser", "Password1", "wrong.email", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Registration_UserAlreadyExists(t *testing.T) {
	repo := new(MockUserRepo)
	service := NewUserService(repo, new(MockJWT), defaultCfg())

	existing := &user.User{Login: "existing"}

	repo.On("GetUser", "existing").Return(existing, nil)

	u, err := service.Registration("existing", "Password1", "email@test.com", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Registration_RepoErrorOnCheck(t *testing.T) {
	repo := new(MockUserRepo)
	service := NewUserService(repo, new(MockJWT), defaultCfg())

	repo.On("GetUser", "erroruser").Return((*user.User)(nil), errors.New("db error"))

	u, err := service.Registration("erroruser", "Password1", "email@test.com", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Registration_RepoSaveError(t *testing.T) {
	repo := new(MockUserRepo)
	service := NewUserService(repo, new(MockJWT), defaultCfg())

	repo.On("GetUser", "newuser").Return((*user.User)(nil), errors.New("user not found"))
	repo.On("SaveUser", mock.Anything).Return(errors.New("save error"))

	u, err := service.Registration("newuser", "Password1", "email@test.com", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

// -------------------- VALIDATORS --------------------

func Test_isValidLogin(t *testing.T) {
	s := NewUserService(nil, nil, defaultCfg())

	assert.Error(t, s.isValidLogin("ab")) // too short
	assert.Error(t, s.isValidLogin("this_user_name_is_way_too_long"))
	assert.Error(t, s.isValidLogin("invalid login")) // space not allowed
	assert.NoError(t, s.isValidLogin("Valid_User"))
}

func Test_isValidPassword(t *testing.T) {
	s := NewUserService(nil, nil, defaultCfg())

	assert.Error(t, s.isValidPassword("short"))
	assert.Error(t, s.isValidPassword("nouppercase1"))
	assert.Error(t, s.isValidPassword("NOLOWER1"))
	assert.Error(t, s.isValidPassword("NoDigitHere"))
	assert.NoError(t, s.isValidPassword("Password1"))
}

func Test_isValidTelegramUsername(t *testing.T) {
	s := NewUserService(nil, nil, defaultCfg())

	assert.Error(t, s.isValidTelegramUsername("abc")) // not digits
	assert.NoError(t, s.isValidTelegramUsername("12345"))
}

func Test_isValidEmail(t *testing.T) {
	s := NewUserService(nil, nil, defaultCfg())

	assert.Error(t, s.isValidEmail("a@b.c"))     // too short
	assert.Error(t, s.isValidEmail("not-email")) // format fail
	assert.NoError(t, s.isValidEmail("test@mail.com"))
}

// -------------------- JWT Passthrough --------------------

func TestUserService_RefreshTokens(t *testing.T) {
	jwt := new(MockJWT)
	service := NewUserService(nil, jwt, defaultCfg())

	jwtResp := &auth.JWTResponse{AccessToken: "a"}
	jwt.On("RefreshTokens", "r").Return(jwtResp, nil)

	res, err := service.RefreshTokens("r")
	assert.NoError(t, err)
	assert.Equal(t, "a", res.AccessToken)
}

func TestUserService_ValidateTokens(t *testing.T) {
	jwt := new(MockJWT)
	service := NewUserService(nil, jwt, defaultCfg())

	payload := &auth.JWTPayload{UserID: "1"}
	jwt.On("ValidateTokens", "t").Return(payload, nil)

	res, err := service.ValidateTokens("t")
	assert.NoError(t, err)
	assert.Equal(t, "1", res.UserID)
}
