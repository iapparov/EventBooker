package user

import (
	"errors"
	"eventbooker/internal/auth"
	"eventbooker/internal/config"
	"eventbooker/internal/domain/user"
	"fmt"
	"regexp"
	"strconv"
	"unicode"
	"unicode/utf8"

	wbzlog "github.com/wb-go/wbf/zlog"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo UserStorageProvider
	jwt JwtAuthProvider
	cfg *config.AppConfig
}

type JwtAuthProvider interface {
	GenerateTokens(user *user.User) (*auth.JWTResponse, error)
	ValidateTokens(tokenStr string) (*auth.JWTPayload, error)
	RefreshTokens(refreshToken string) (*auth.JWTResponse, error)
}

type UserStorageProvider interface{
	GetUser(login string) (*user.User, error)
	SaveUser(user *user.User) error
}

func NewUserService (repo UserStorageProvider, jwt JwtAuthProvider, cfg *config.AppConfig) *UserService {
	return &UserService{
		repo: repo,
		jwt: jwt,
		cfg: cfg,
	}
}

func (s *UserService) Login(Login, Password string) (*auth.JWTResponse, error){
	if Login == "" || Password == "" {
		wbzlog.Logger.Debug().Msg("login or password cant be empty")
		return nil , errors.New("login or password cant be empty")
	}

	user, err := s.repo.GetUser(Login)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(Password))
	if err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid password")
		return nil, err
	}

	jwtresp, err := s.jwt.GenerateTokens(user)
	if err != nil {
		return nil, err
	}
	return jwtresp, nil
}

func (s *UserService) Registration(Login, Password, Email, Telegram string) (*user.User, error){
	if err := s.isValidLogin(Login); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid login")
		return nil, err
	}

	if err := s.isValidPassword(Password); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid password")
		return nil, err
	}

	if err := s.isValidTelegramUsername(Telegram); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid telegram")
		return nil, err
	}

	if err := s.isValidEmail(Email); err != nil {
		wbzlog.Logger.Debug().Err(err).Msg("invalid email")
		return nil, err
	}

	ch, err := s.repo.GetUser(Login)
	if err != nil && err.Error() != "user not found" {
		wbzlog.Logger.Error().Err(err).Msg("cant check existing user")
		return nil, err
	}
	
	if ch != nil {
		wbzlog.Logger.Debug().Msg("user with this login already exists")
		return nil, errors.New("user with this login already exists")
	}

	user, err := user.NewUser(Login, Password, Email, Telegram)

	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("cant create new user")
		return nil, err
	}

	err = s.repo.SaveUser(user)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) RefreshTokens(refreshToken string) (*auth.JWTResponse, error) {
	return s.jwt.RefreshTokens(refreshToken)
}

func (s *UserService) ValidateTokens(tokenStr string) (*auth.JWTPayload, error) {
	return s.jwt.ValidateTokens(tokenStr)
}

func (s *UserService) isValidLogin(login string) (error) {
	if utf8.RuneCountInString(login) < s.cfg.UserConfig.MinLength || utf8.RuneCountInString(login) > s.cfg.UserConfig.MaxLength {
		return fmt.Errorf("invalid login length . Must be between %d and %d characters", s.cfg.UserConfig.MinLength, s.cfg.UserConfig.MaxLength)
	}

	escapedChars := regexp.QuoteMeta(s.cfg.UserConfig.AllowedCharacters)
	loginRegexp := regexp.MustCompile(`^[` + escapedChars + `]+$`)
	if !loginRegexp.MatchString(login) {
		return errors.New("invalid login characters. Must contain only letters, digits, underscores, or hyphens and must not contain spaces")
	}
	return nil
}

func (s *UserService) isValidPassword(password string) error {
	cfg := s.cfg.PasswordConfig

	l := utf8.RuneCountInString(password)
	if l < cfg.MinLength || l > cfg.MaxLength {
		return fmt.Errorf(
			"invalid password length: must be %d–%d characters",
			cfg.MinLength, cfg.MaxLength,
		)
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

func (s *UserService) isValidTelegramUsername(username string) error {
	// l := utf8.RuneCountInString(username)
	// if l < 5|| l > 32 { // 5 и 32 это стандарт телеграмма
	// 	return fmt.Errorf("telegram username length must be between %d and %d characters", 5, 32)
	// }

	// // Telegram разрешает a-zA-Z0-9 и _
	// re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	// if !re.MatchString(username) {
	// 	return errors.New("invalid telegram username format: must start with a letter or underscore and contain only letters, digits or underscores")
	// }

	_, err := strconv.Atoi(username)
	if err != nil {
		return errors.New("telegram chat_id must be digit")
	}

	return nil
}

func (s *UserService) isValidEmail(email string) error {
	l := utf8.RuneCountInString(email)
	if l < 6 { // 1@1.ru от 6 символов
		return errors.New("email length must be bigger than 6 characters")
	}

	
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`) // для базовой проверки формата asda@asdasd.asdads
	if !re.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}