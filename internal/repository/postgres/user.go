package postgres

import (
	"context"
	"database/sql"
	"errors"

	"eventbooker/internal/domain/user"

	"github.com/wb-go/wbf/retry"
	wbzlog "github.com/wb-go/wbf/zlog"
)

// GetUser retrieves a user by login.
func (r *Repository) GetUser(ctx context.Context, login string) (*user.User, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `SELECT id, login, password, created_at, email, telegram FROM users WHERE login = $1`

	row, err := r.db.QueryRowWithRetry(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, query, login)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to execute get user query")
		return nil, err
	}

	var u user.User
	err = row.Scan(&u.ID, &u.Login, &u.Password, &u.CreatedAt, &u.Email, &u.Telegram)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to scan user row")
		return nil, err
	}

	return &u, nil
}

// GetUserByUUID retrieves a user by UUID.
func (r *Repository) GetUserByUUID(ctx context.Context, id string) (*user.User, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `SELECT id, login, password, created_at, email, telegram FROM users WHERE id = $1`

	row, err := r.db.QueryRowWithRetry(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, query, id)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to execute get user query")
		return nil, err
	}

	var u user.User
	err = row.Scan(&u.ID, &u.Login, &u.Password, &u.CreatedAt, &u.Email, &u.Telegram)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to scan user row")
		return nil, err
	}

	return &u, nil
}

// SaveUser inserts a new user.
func (r *Repository) SaveUser(ctx context.Context, u *user.User) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `INSERT INTO users (id, login, password, created_at, email, telegram) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecWithRetry(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, query,
		u.ID, u.Login, u.Password, u.CreatedAt, u.Email, u.Telegram,
	)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to insert user")
		return err
	}

	return nil
}
