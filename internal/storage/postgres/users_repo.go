package postgres

import (
	"context"
	"database/sql"
	"eventbooker/internal/domain/user"
	"errors"
	"github.com/wb-go/wbf/retry"
	wbzlog "github.com/wb-go/wbf/zlog"
)



func (p *Postgres) GetUser(login string) (*user.User, error) {
	ctx := context.Background()

	query := `
		SELECT id, login, password, created_at, email, telegram
		FROM users
		WHERE login = $1
	`
	var u user.User
	row, err := p.db.QueryRowWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query, login)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Failed to execute get user query")
		return nil, err
	}
	
	err = row.Scan(
		&u.Id,
		&u.Login,
		&u.Password,
		&u.CreatedAt,
		&u.Email,
		&u.Telegram,
	)
	if err != nil && err != sql.ErrNoRows {
		wbzlog.Logger.Error().Err(err).Msg("Failed to scan user row")
		return nil, err
	}

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	return &u, nil
}
func (p *Postgres) SaveUser(user *user.User) error{
	ctx := context.Background()

	query := `
		INSERT INTO users (id, login, password, created_at, email, telegram)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := p.db.ExecWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query,
		user.Id,
		user.Login,
		user.Password,
		user.CreatedAt,
		user.Email,
		user.Telegram,
	)

	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Failed to execute insert user query")
		return err
	}
	return nil
}

