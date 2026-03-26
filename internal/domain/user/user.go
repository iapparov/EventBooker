package user

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User is the domain model for a user account.
type User struct {
	ID        uuid.UUID
	Login     string
	Password  []byte
	CreatedAt time.Time
	Email     string
	Telegram  string
}

// New creates a new User with a hashed password.
func New(login, password, email, telegram string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        uuid.New(),
		Login:     login,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		Email:     email,
		Telegram:  telegram,
	}, nil
}
