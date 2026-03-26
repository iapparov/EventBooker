package service

import (
	"context"
	"errors"
	"testing"

	"eventbooker/internal/auth"
	"eventbooker/internal/config"
	"eventbooker/internal/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) GetUser(ctx context.Context, login string) (*user.User, error) {
	args := m.Called(login)
	return args.Get(0).(*user.User), args.Error(1)
}
func (m *mockUserRepo) SaveUser(ctx context.Context, u *user.User) error {
	return m.Called(u).Error(0)
}

type mockJWT struct{ mock.Mock }

func (m *mockJWT) GenerateTokens(u *user.User) (*auth.Response, error) {
	args := m.Called(u)
	return args.Get(0).(*auth.Response), args.Error(1)
}
func (m *mockJWT) ValidateToken(t string) (*auth.Payload, error) {
	args := m.Called(t)
	return args.Get(0).(*auth.Payload), args.Error(1)
}
func (m *mockJWT) RefreshTokens(r string) (*auth.Response, error) {
	args := m.Called(r)
	return args.Get(0).(*auth.Response), args.Error(1)
}

func defaultUserCfg() *config.AppConfig {
	return &config.AppConfig{
		User: config.UserConfig{
			MinLength:         3,
			MaxLength:         20,
			AllowedCharacters: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-",
		},
		Password: config.PasswordConfig{
			MinLength:    6,
			MaxLength:    32,
			RequireUpper: true,
			RequireLower: true,
			RequireDigit: true,
		},
	}
}

func TestUserService_Login_Success(t *testing.T) {
	repo := new(mockUserRepo)
	jwt := new(mockJWT)
	svc := NewUserService(repo, jwt, defaultUserCfg())

	hash, _ := bcrypt.GenerateFromPassword([]byte("Password1"), bcrypt.DefaultCost)
	u := &user.User{Login: "test", Password: hash}
	repo.On("GetUser", "test").Return(u, nil)
	jwtResp := &auth.Response{AccessToken: "token"}
	jwt.On("GenerateTokens", u).Return(jwtResp, nil)

	resp, err := svc.Login(context.Background(), "test", "Password1")
	assert.NoError(t, err)
	assert.Equal(t, "token", resp.AccessToken)
	repo.AssertExpectations(t)
	jwt.AssertExpectations(t)
}

func TestUserService_Login_EmptyFields(t *testing.T) {
	svc := NewUserService(new(mockUserRepo), new(mockJWT), defaultUserCfg())
	resp, err := svc.Login(context.Background(), "", "pass")
	assert.Error(t, err)
	assert.Nil(t, resp)
	resp2, err2 := svc.Login(context.Background(), "login", "")
	assert.Error(t, err2)
	assert.Nil(t, resp2)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewUserService(repo, new(mockJWT), defaultUserCfg())
	repo.On("GetUser", "test").Return(&user.User{}, errors.New("user not found"))
	resp, err := svc.Login(context.Background(), "test", "pass")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUserService_Login_InvalidPassword(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewUserService(repo, new(mockJWT), defaultUserCfg())
	hash, _ := bcrypt.GenerateFromPassword([]byte("Password1"), bcrypt.DefaultCost)
	u := &user.User{Password: hash}
	repo.On("GetUser", "test").Return(u, nil)
	resp, err := svc.Login(context.Background(), "test", "wrong")
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestUserService_Register_Success(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewUserService(repo, new(mockJWT), defaultUserCfg())
	repo.On("GetUser", "newuser").Return((*user.User)(nil), errors.New("user not found"))
	repo.On("SaveUser", mock.Anything).Return(nil)
	u, err := svc.Register(context.Background(), "newuser", "Password1", "email@test.com", "12345")
	assert.NoError(t, err)
	assert.Equal(t, "newuser", u.Login)
	repo.AssertExpectations(t)
}

func TestUserService_Register_InvalidLogin(t *testing.T) {
	svc := NewUserService(new(mockUserRepo), new(mockJWT), defaultUserCfg())
	u, err := svc.Register(context.Background(), "ab", "Password1", "email@test.com", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Register_InvalidPassword(t *testing.T) {
	svc := NewUserService(new(mockUserRepo), new(mockJWT), defaultUserCfg())
	u, err := svc.Register(context.Background(), "validUser", "short", "email@test.com", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Register_InvalidTelegram(t *testing.T) {
	svc := NewUserService(new(mockUserRepo), new(mockJWT), defaultUserCfg())
	u, err := svc.Register(context.Background(), "validUser", "Password1", "email@test.com", "abc")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Register_InvalidEmail(t *testing.T) {
	svc := NewUserService(new(mockUserRepo), new(mockJWT), defaultUserCfg())
	u, err := svc.Register(context.Background(), "validUser", "Password1", "wrong.email", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Register_UserAlreadyExists(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewUserService(repo, new(mockJWT), defaultUserCfg())
	existing := &user.User{Login: "existing"}
	repo.On("GetUser", "existing").Return(existing, nil)
	u, err := svc.Register(context.Background(), "existing", "Password1", "email@test.com", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Register_RepoErrorOnCheck(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewUserService(repo, new(mockJWT), defaultUserCfg())
	repo.On("GetUser", "erroruser").Return((*user.User)(nil), errors.New("db error"))
	u, err := svc.Register(context.Background(), "erroruser", "Password1", "email@test.com", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestUserService_Register_RepoSaveError(t *testing.T) {
	repo := new(mockUserRepo)
	svc := NewUserService(repo, new(mockJWT), defaultUserCfg())
	repo.On("GetUser", "newuser").Return((*user.User)(nil), errors.New("user not found"))
	repo.On("SaveUser", mock.Anything).Return(errors.New("save error"))
	u, err := svc.Register(context.Background(), "newuser", "Password1", "email@test.com", "12345")
	assert.Error(t, err)
	assert.Nil(t, u)
}

func Test_validateLogin(t *testing.T) {
	svc := NewUserService(nil, nil, defaultUserCfg())
	assert.Error(t, svc.validateLogin("ab"))
	assert.Error(t, svc.validateLogin("this_user_name_is_way_too_long"))
	assert.Error(t, svc.validateLogin("invalid login"))
	assert.NoError(t, svc.validateLogin("Valid_User"))
}

func Test_validatePassword(t *testing.T) {
	svc := NewUserService(nil, nil, defaultUserCfg())
	assert.Error(t, svc.validatePassword("short"))
	assert.Error(t, svc.validatePassword("nouppercase1"))
	assert.Error(t, svc.validatePassword("NOLOWER1"))
	assert.Error(t, svc.validatePassword("NoDigitHere"))
	assert.NoError(t, svc.validatePassword("Password1"))
}

func Test_validateTelegram(t *testing.T) {
	svc := NewUserService(nil, nil, defaultUserCfg())
	assert.Error(t, svc.validateTelegram("abc"))
	assert.NoError(t, svc.validateTelegram("12345"))
}

func Test_validateEmail(t *testing.T) {
	svc := NewUserService(nil, nil, defaultUserCfg())
	assert.Error(t, svc.validateEmail("a@b.c"))
	assert.Error(t, svc.validateEmail("not-email"))
	assert.NoError(t, svc.validateEmail("test@mail.com"))
}

func TestUserService_RefreshTokens(t *testing.T) {
	jwt := new(mockJWT)
	svc := NewUserService(nil, jwt, defaultUserCfg())
	jwtResp := &auth.Response{AccessToken: "a"}
	jwt.On("RefreshTokens", "r").Return(jwtResp, nil)
	res, err := svc.RefreshTokens("r")
	assert.NoError(t, err)
	assert.Equal(t, "a", res.AccessToken)
}

func TestUserService_ValidateToken(t *testing.T) {
	jwt := new(mockJWT)
	svc := NewUserService(nil, jwt, defaultUserCfg())
	payload := &auth.Payload{UserID: "1"}
	jwt.On("ValidateToken", "t").Return(payload, nil)
	res, err := svc.ValidateToken("t")
	assert.NoError(t, err)
	assert.Equal(t, "1", res.UserID)
}
