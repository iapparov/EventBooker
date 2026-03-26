package user_test

import (
	"testing"
	"time"

	u "eventbooker/internal/domain/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestNew_Success(t *testing.T) {
	login := "testuser"
	password := "MyPassword123"
	email := "test@example.com"
	telegram := "@test"

	user, err := u.New(login, password, email, telegram)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("user is nil")
	}
	if user.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if user.Login != login {
		t.Errorf("expected login %s, got %s", login, user.Login)
	}
	if user.Email != email {
		t.Errorf("expected email %s, got %s", email, user.Email)
	}
	if user.Telegram != telegram {
		t.Errorf("expected telegram %s, got %s", telegram, user.Telegram)
	}
	if time.Since(user.CreatedAt) > time.Second {
		t.Error("CreatedAt should be set to current time")
	}
	if string(user.Password) == password {
		t.Error("password must be hashed, not equal to raw password")
	}
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		t.Errorf("bcrypt hash does not match the original password: %v", err)
	}
}

func TestNew_HashError(t *testing.T) {
	veryLongPassword := make([]byte, 10000)
	for i := range veryLongPassword {
		veryLongPassword[i] = 'A'
	}
	_, err := u.New("user", string(veryLongPassword), "e@mail.com", "@tg")
	if err == nil {
		t.Fatal("expected error from bcrypt, got nil")
	}
}
