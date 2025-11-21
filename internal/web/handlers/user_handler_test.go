package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"eventbooker/internal/auth"
	"eventbooker/internal/domain/user"
	"eventbooker/internal/web/dto"
	"eventbooker/internal/web/handlers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------- MOCKS --------------------

type MockUserService struct {
	LoginFn          func(login, password string) (*auth.JWTResponse, error)
	RegistrationFn   func(login, password, email, telegram string) (*user.User, error)
	RefreshTokensFn  func(tokenStr string) (*auth.JWTResponse, error)
	ValidateTokensFn func(tokenStr string) (*auth.JWTPayload, error)
}

func (m *MockUserService) Login(login, password string) (*auth.JWTResponse, error) {
	return m.LoginFn(login, password)
}

func (m *MockUserService) Registration(login, password, email, telegram string) (*user.User, error) {
	return m.RegistrationFn(login, password, email, telegram)
}

func (m *MockUserService) RefreshTokens(tokenStr string) (*auth.JWTResponse, error) {
	return m.RefreshTokensFn(tokenStr)
}

func (m *MockUserService) ValidateTokens(tokenStr string) (*auth.JWTPayload, error) {
	return m.ValidateTokensFn(tokenStr)
}

// ---------------- UTILS --------------------

func performRequestUser(hf func(*gin.Context), method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	hf(c)
	return w
}

// ---------------- TESTS --------------------

func TestUserHandler_RegisterUser_Success(t *testing.T) {
	mockService := &MockUserService{
		RegistrationFn: func(login, password, email, telegram string) (*user.User, error) {
			return &user.User{
				Id:       uuid.New(),
				Login:    login,
				Email:    email,
				Telegram: telegram,
			}, nil
		},
	}
	h := handlers.NewUserHandler(mockService)

	req := dto.UserRegistrationRequest{
		Login:    "testuser",
		Password: "password123",
		Email:    "test@test.com",
		Telegram: "tguser",
	}

	w := performRequestUser(h.RegisterUser, "POST", "/register", req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserHandler_RegisterUser_InvalidJSON(t *testing.T) {
	h := handlers.NewUserHandler(&MockUserService{})

	w := performRequestUser(h.RegisterUser, "POST", "/register", "{bad json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_RegisterUser_ServiceError(t *testing.T) {
	mockService := &MockUserService{
		RegistrationFn: func(login, password, email, telegram string) (*user.User, error) {
			return nil, errors.New("service error")
		},
	}
	h := handlers.NewUserHandler(mockService)

	req := dto.UserRegistrationRequest{
		Login:    "testuser",
		Password: "password123",
		Email:    "test@test.com",
		Telegram: "tguser",
	}

	w := performRequestUser(h.RegisterUser, "POST", "/register", req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUserHandler_LoginUser_Success(t *testing.T) {
	mockService := &MockUserService{
		LoginFn: func(login, password string) (*auth.JWTResponse, error) {
			return &auth.JWTResponse{
				AccessToken:  "access123",
				RefreshToken: "refresh123",
			}, nil
		},
	}
	h := handlers.NewUserHandler(mockService)

	req := dto.UserLoginRequest{
		Login:    "testuser",
		Password: "password123",
	}

	w := performRequestUser(h.LoginUser, "POST", "/login", req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserHandler_LoginUser_InvalidJSON(t *testing.T) {
	h := handlers.NewUserHandler(&MockUserService{})

	w := performRequestUser(h.LoginUser, "POST", "/login", "{bad json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_LoginUser_Unauthorized(t *testing.T) {
	mockService := &MockUserService{
		LoginFn: func(login, password string) (*auth.JWTResponse, error) {
			return nil, errors.New("invalid credentials")
		},
	}
	h := handlers.NewUserHandler(mockService)

	req := dto.UserLoginRequest{
		Login:    "testuser",
		Password: "wrongpass",
	}

	w := performRequestUser(h.LoginUser, "POST", "/login", req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestUserHandler_RefreshToken_Success(t *testing.T) {
	mockService := &MockUserService{
		RefreshTokensFn: func(tokenStr string) (*auth.JWTResponse, error) {
			return &auth.JWTResponse{
				AccessToken:  "newaccess",
				RefreshToken: "newrefresh",
			}, nil
		},
	}
	h := handlers.NewUserHandler(mockService)

	req := dto.TokenRefreshRequest{
		RefreshToken: "refresh123",
	}

	w := performRequestUser(h.RefreshToken, "POST", "/refresh", req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserHandler_RefreshToken_InvalidJSON(t *testing.T) {
	h := handlers.NewUserHandler(&MockUserService{})

	w := performRequestUser(h.RefreshToken, "POST", "/refresh", "{bad json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_RefreshToken_Unauthorized(t *testing.T) {
	mockService := &MockUserService{
		RefreshTokensFn: func(tokenStr string) (*auth.JWTResponse, error) {
			return nil, errors.New("invalid token")
		},
	}
	h := handlers.NewUserHandler(mockService)

	req := dto.TokenRefreshRequest{
		RefreshToken: "badtoken",
	}

	w := performRequestUser(h.RefreshToken, "POST", "/refresh", req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
