package auth_test

import (
	"testing"
	"time"

	"eventbooker/internal/auth"
	"eventbooker/internal/config"
	"eventbooker/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func newTestJWT() *auth.Service {
	return auth.NewService(&config.JWTConfig{
		AccessSecret:    "access-secret",
		RefreshSecret:   "refresh-secret",
		ExpAccessToken:  1,
		ExpRefreshToken: 1,
	})
}

func newTestUser() *user.User {
	return &user.User{ID: uuid.New()}
}

func TestGenerateTokens_Success(t *testing.T) {
	s := newTestJWT()
	u := newTestUser()

	resp, err := s.GenerateTokens(u)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatal("tokens must not be empty")
	}
}

func TestValidateToken_Success(t *testing.T) {
	s := newTestJWT()
	u := newTestUser()
	tokens, _ := s.GenerateTokens(u)

	payload, err := s.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if payload.UserID != u.ID.String() {
		t.Fatalf("invalid payload user ID")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	s := newTestJWT()
	_, err := s.ValidateToken("this_is_invalid_token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateToken_NoUUIDField(t *testing.T) {
	s := newTestJWT()
	claims := jwt.MapClaims{"exp": time.Now().Add(time.Minute).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("access-secret"))

	_, err := s.ValidateToken(tokenStr)
	if err == nil {
		t.Fatal("expected error because uuid field is missing")
	}
}

func TestRefreshTokens_Success(t *testing.T) {
	s := newTestJWT()
	u := newTestUser()
	tokens, _ := s.GenerateTokens(u)

	newTokens, err := s.RefreshTokens(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if newTokens.AccessToken == "" || newTokens.RefreshToken == "" {
		t.Fatal("new tokens must not be empty")
	}
}

func TestRefreshTokens_Invalid(t *testing.T) {
	s := newTestJWT()
	_, err := s.RefreshTokens("invalid_refresh_token")
	if err == nil {
		t.Fatal("expected error for invalid refresh token")
	}
}

func TestRefreshTokens_NoUUIDField(t *testing.T) {
	s := newTestJWT()
	claims := jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("refresh-secret"))

	_, err := s.RefreshTokens(tokenStr)
	if err == nil {
		t.Fatal("expected error because uuid missing")
	}
}

func TestValidateAccessToken_Expired(t *testing.T) {
	cfg := &config.JWTConfig{
		AccessSecret:   "access-secret",
		RefreshSecret:  "refresh-secret",
		ExpAccessToken: -1,
	}

	s := auth.NewService(cfg)
	u := newTestUser()
	expiredToken, _ := s.GenerateTokens(u)

	_, err := s.ValidateToken(expiredToken.AccessToken)
	if err == nil {
		t.Fatal("expected error: access token expired")
	}
}

func TestValidateRefreshToken_Expired(t *testing.T) {
	cfg := &config.JWTConfig{
		AccessSecret:    "access-secret",
		RefreshSecret:   "refresh-secret",
		ExpRefreshToken: -1,
	}

	s := auth.NewService(cfg)
	u := newTestUser()
	tokens, _ := s.GenerateTokens(u)

	_, err := s.RefreshTokens(tokens.RefreshToken)
	if err == nil {
		t.Fatal("expected error: refresh token expired")
	}
}
