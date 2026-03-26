package auth

import (
	"errors"
	"time"

	"eventbooker/internal/config"
	"eventbooker/internal/domain/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Response holds a pair of JWT tokens.
type Response struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresIn  int64
	RefreshExpiresIn int64
	TokenType        string
}

// Payload holds the validated token claims.
type Payload struct {
	UserID string
}

// Service handles JWT token operations.
type Service struct {
	accessSecret    string
	refreshSecret   string
	expAccessToken  int // minutes
	expRefreshToken int // hours
}

// NewService creates a new JWT Service.
func NewService(cfg *config.JWTConfig) *Service {
	return &Service{
		accessSecret:    cfg.AccessSecret,
		refreshSecret:   cfg.RefreshSecret,
		expAccessToken:  cfg.ExpAccessToken,
		expRefreshToken: cfg.ExpRefreshToken,
	}
}

// GenerateTokens creates a new access/refresh token pair.
func (s *Service) GenerateTokens(u *user.User) (*Response, error) {
	access, err := s.generateAccessToken(u)
	if err != nil {
		return nil, err
	}

	refresh, err := s.generateRefreshToken(u)
	if err != nil {
		return nil, err
	}

	return &Response{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

// ValidateToken validates an access token and returns its payload.
func (s *Service) ValidateToken(tokenStr string) (*Payload, error) {
	claims, err := s.validateAccessToken(tokenStr)
	if err != nil {
		return nil, err
	}

	uuidStr, ok := claims["uuid"].(string)
	if !ok {
		return nil, errors.New("invalid token payload")
	}

	return &Payload{UserID: uuidStr}, nil
}

// RefreshTokens generates a new token pair from a valid refresh token.
func (s *Service) RefreshTokens(refreshToken string) (*Response, error) {
	claims, err := s.validateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	uuidStr, ok := claims["uuid"].(string)
	if !ok {
		return nil, errors.New("invalid refresh token payload")
	}

	u := &user.User{ID: uuid.MustParse(uuidStr)}
	return s.GenerateTokens(u)
}

func (s *Service) generateAccessToken(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"uuid": u.ID.String(),
		"exp":  time.Now().Add(time.Minute * time.Duration(s.expAccessToken)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.accessSecret))
}

func (s *Service) generateRefreshToken(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"uuid": u.ID.String(),
		"exp":  time.Now().Add(time.Hour * time.Duration(s.expRefreshToken)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.refreshSecret))
}

func (s *Service) validateAccessToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.accessSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid access token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	if exp, ok := claims["exp"].(float64); !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

func (s *Service) validateRefreshToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.refreshSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	if exp, ok := claims["exp"].(float64); !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		return nil, errors.New("refresh token has expired")
	}

	return claims, nil
}
