package user_test

import (
	u "eventbooker/internal/domain/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func TestNewUser_Success(t *testing.T) {
	login := "testuser"
	password := "MyPassword123"
	email := "test@example.com"
	telegram := "@test"

	// --- WHEN ---
	user, err := u.NewUser(login, password, email, telegram)

	// --- THEN ---
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("user is nil")
	}

	// Check UUID
	if user.Id == uuid.Nil {
		t.Error("expected non-nil UUID")
	}

	// Check login
	if user.Login != login {
		t.Errorf("expected login %s, got %s", login, user.Login)
	}

	// Check email
	if user.Email != email {
		t.Errorf("expected email %s, got %s", email, user.Email)
	}

	// Check telegram
	if user.Telegram != telegram {
		t.Errorf("expected telegram %s, got %s", telegram, user.Telegram)
	}

	// Check CreatedAt
	if time.Since(user.CreatedAt) > time.Second {
		t.Error("CreatedAt should be set to current time")
	}

	// Password must be hashed
	if string(user.Password) == password {
		t.Error("password must be hashed, not equal to raw password")
	}

	// Check bcrypt validity
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		t.Errorf("bcrypt hash does not match the original password: %v", err)
	}
}

func TestNewUser_HashError(t *testing.T) {
	// bcrypt errors occur only on system-level failures,
	// but we can simulate by passing extremely long password.

	// Create an unrealistically huge password to force bcrypt error
	// bcrypt's max password length is ~72 bytes; we exceed it intentionally
	veryLongPassword := make([]byte, 10000)
	for i := range veryLongPassword {
		veryLongPassword[i] = 'A'
	}

	_, err := u.NewUser("user", string(veryLongPassword), "e@mail.com", "@tg")

	if err == nil {
		t.Fatal("expected error from bcrypt, got nil")
	}
}
