package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"eventbooker/internal/auth"
	"eventbooker/internal/domain/user"
	"eventbooker/internal/transport/http/dto"
	"eventbooker/internal/transport/http/handler"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockUserService struct {
	LoginFn         func(ctx context.Context, login, password string) (*auth.Response, error)
	RegisterFn      func(ctx context.Context, login, password, email, telegram string) (*user.User, error)
	RefreshTokensFn func(tokenStr string) (*auth.Response, error)
	ValidateTokenFn func(tokenStr string) (*auth.Payload, error)
}

func (m *mockUserService) Login(ctx context.Context, login, password string) (*auth.Response, error) {
	return m.LoginFn(ctx, login, password)
}
func (m *mockUserService) Register(ctx context.Context, login, password, email, telegram string) (*user.User, error) {
	return m.RegisterFn(ctx, login, password, email, telegram)
}
func (m *mockUserService) RefreshTokens(tokenStr string) (*auth.Response, error) {
	return m.RefreshTokensFn(tokenStr)
}
func (m *mockUserService) ValidateToken(tokenStr string) (*auth.Payload, error) {
	return m.ValidateTokenFn(tokenStr)
}

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

func TestUserHandler_RegisterUser_Success(t *testing.T) {
	mock := &mockUserService{
		RegisterFn: func(ctx context.Context, login, password, email, telegram string) (*user.User, error) {
			return &user.User{ID: uuid.New(), Login: login, Email: email, Telegram: telegram}, nil
		},
	}
	h := handler.NewUserHandler(mock)
	req := dto.UserRegistrationRequest{Login: "testuser", Password: "password123", Email: "test@test.com", Telegram: "tguser"}
	w := performRequestUser(h.RegisterUser, "POST", "/register", req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserHandler_RegisterUser_InvalidJSON(t *testing.T) {
	h := handler.NewUserHandler(&mockUserService{})
	w := performRequestUser(h.RegisterUser, "POST", "/register", "{bad json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_RegisterUser_ServiceError(t *testing.T) {
	mock := &mockUserService{
		RegisterFn: func(ctx context.Context, login, password, email, telegram string) (*user.User, error) {
			return nil, errors.New("service error")
		},
	}
	h := handler.NewUserHandler(mock)
	req := dto.UserRegistrationRequest{Login: "testuser", Password: "password123", Email: "test@test.com", Telegram: "tguser"}
	w := performRequestUser(h.RegisterUser, "POST", "/register", req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUserHandler_LoginUser_Success(t *testing.T) {
	mock := &mockUserService{
		LoginFn: func(ctx context.Context, login, password string) (*auth.Response, error) {
			return &auth.Response{AccessToken: "access123", RefreshToken: "refresh123"}, nil
		},
	}
	h := handler.NewUserHandler(mock)
	req := dto.UserLoginRequest{Login: "testuser", Password: "password123"}
	w := performRequestUser(h.LoginUser, "POST", "/login", req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserHandler_LoginUser_InvalidJSON(t *testing.T) {
	h := handler.NewUserHandler(&mockUserService{})
	w := performRequestUser(h.LoginUser, "POST", "/login", "{bad json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_LoginUser_Unauthorized(t *testing.T) {
	mock := &mockUserService{
		LoginFn: func(ctx context.Context, login, password string) (*auth.Response, error) {
			return nil, errors.New("invalid credentials")
		},
	}
	h := handler.NewUserHandler(mock)
	req := dto.UserLoginRequest{Login: "testuser", Password: "wrongpass"}
	w := performRequestUser(h.LoginUser, "POST", "/login", req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestUserHandler_RefreshToken_Success(t *testing.T) {
	mock := &mockUserService{
		RefreshTokensFn: func(tokenStr string) (*auth.Response, error) {
			return &auth.Response{AccessToken: "newaccess", RefreshToken: "newrefresh"}, nil
		},
	}
	h := handler.NewUserHandler(mock)
	req := dto.TokenRefreshRequest{RefreshToken: "refresh123"}
	w := performRequestUser(h.RefreshToken, "POST", "/refresh", req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserHandler_RefreshToken_InvalidJSON(t *testing.T) {
	h := handler.NewUserHandler(&mockUserService{})
	w := performRequestUser(h.RefreshToken, "POST", "/refresh", "{bad json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_RefreshToken_Unauthorized(t *testing.T) {
	mock := &mockUserService{
		RefreshTokensFn: func(tokenStr string) (*auth.Response, error) {
			return nil, errors.New("invalid token")
		},
	}
	h := handler.NewUserHandler(mock)
	req := dto.TokenRefreshRequest{RefreshToken: "badtoken"}
	w := performRequestUser(h.RefreshToken, "POST", "/refresh", req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
