package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"unicode"
	"unicode/utf8"

	"eventbooker/internal/auth"
	"eventbooker/internal/config"
	"eventbooker/internal/domain/user"

	wbzlog "github.com/wb-go/wbf/zlog"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository defines the storage operations needed by UserService.
type UserRepository interface {
	GetUser(ctx context.Context, login string) (*user.User, error)
	SaveUser(ctx context.Context, u *user.User) error
}

// TokenProvider defines the JWT operations needed by UserService.
type TokenProvider interface {
	GenerateTokens(u *user.User) (*auth.Response, error)
	ValidateToken(tokenStr string) (*auth.Payload, error)
	RefreshTokens(refreshToken string) (*auth.Response, error)
}

// UserService handles user business logic.
type UserService struct {
	repo UserRepository
	jwt  TokenProvider
	cfg  *config.AppConfig
}

// NewUserService creates a new UserService.
func NewUserService(repo UserRepository, jwt TokenProvider, cfg *config.AppConfig) *UserService {
	return &UserService{
		repo: repo,
		jwt:  jwt,
		cfg:  cfg,
	}
}

// Login authenticates a user and returns JWT tokens.
func (s *UserService) Login(ctx context.Context, login, password string) (*auth.Response, error) {
	if login == "" || password == "" {
		wbzlog.Logger.Debug().Msg("login or password cannot be empty")
		return nil, errors.New("login or password cannot be empty")
	}

	u, err := s.repo.GetUser(ctx, login)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid password")
		return nil, err
	}

	return s.jwt.GenerateTokens(u)
}

// Register creates a new user account.
func (s *UserService) Register(ctx context.Context, login, password, email, telegram string) (*user.User, error) {
	if err := s.validateLogin(login); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid login")
		return nil, err
	}

	if err := s.validatePassword(password); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid password")
		return nil, err
	}

	if err := s.validateTelegram(telegram); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid telegram")
		return nil, err
	}

	if err := s.validateEmail(email); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid email")
		return nil, err
	}

	existing, err := s.repo.GetUser(ctx, login)
	if err != nil && err.Error() != "user not found" {
		wbzlog.Logger.Error().Err(err).Msg("cannot check existing user")
		return nil, err
	}

	if existing != nil {
		wbzlog.Logger.Debug().Msg("user with this login already exists")
		return nil, errors.New("user with this login already exists")
	}

	u, err := user.New(login, password, email, telegram)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("cannot create new user")
		return nil, err
	}

	if err = s.repo.SaveUser(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

// RefreshTokens refreshes JWT tokens.
func (s *UserService) RefreshTokens(refreshToken string) (*auth.Response, error) {
	return s.jwt.RefreshTokens(refreshToken)
}

// ValidateToken validates a JWT token.
func (s *UserService) ValidateToken(tokenStr string) (*auth.Payload, error) {
	return s.jwt.ValidateToken(tokenStr)
}

func (s *UserService) validateLogin(login string) error {
	l := utf8.RuneCountInString(login)
	if l < s.cfg.User.MinLength || l > s.cfg.User.MaxLength {
		return fmt.Errorf("login length must be between %d and %d characters", s.cfg.User.MinLength, s.cfg.User.MaxLength)
	}

	escapedChars := regexp.QuoteMeta(s.cfg.User.AllowedCharacters)
	loginRegexp := regexp.MustCompile(`^[` + escapedChars + `]+$`)
	if !loginRegexp.MatchString(login) {
		return errors.New("login contains invalid characters")
	}

	return nil
}

func (s *UserService) validatePassword(password string) error {
	cfg := s.cfg.Password
	l := utf8.RuneCountInString(password)

	if l < cfg.MinLength || l > cfg.MaxLength {
		return fmt.Errorf("password length must be %d–%d characters", cfg.MinLength, cfg.MaxLength)
	}

	if !utf8.ValidString(password) {
		return errors.New("password contains invalid UTF-8 characters")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if cfg.RequireUpper && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if cfg.RequireLower && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if cfg.RequireDigit && !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	return nil
}

func (s *UserService) validateTelegram(username string) error {
	if _, err := strconv.Atoi(username); err != nil {
		return errors.New("telegram chat_id must be a number")
	}
	return nil
}

func (s *UserService) validateEmail(email string) error {
	if utf8.RuneCountInString(email) < 6 {
		return errors.New("email length must be at least 6 characters")
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !re.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}
