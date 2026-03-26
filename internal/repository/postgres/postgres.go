package postgres

import (
	"context"
	"fmt"
	"time"

	"eventbooker/internal/config"

	wbdb "github.com/wb-go/wbf/dbpg"
	wbzlog "github.com/wb-go/wbf/zlog"
)

// Repository wraps a PostgreSQL database connection pool.
type Repository struct {
	db           *wbdb.DB
	retry        *config.RetryConfig
	queryTimeout time.Duration
}

// withTimeout wraps the incoming context with the configured query timeout.
func (r *Repository) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.queryTimeout)
}

// New creates a new PostgreSQL Repository.
func New(cfg *config.AppConfig) (*Repository, error) {
	masterDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Master.Host,
		cfg.DB.Master.Port,
		cfg.DB.Master.User,
		cfg.DB.Master.Password,
		cfg.DB.Master.DBName,
	)

	slaveDSNs := make([]string, 0, len(cfg.DB.Slaves))
	for _, slave := range cfg.DB.Slaves {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			slave.Host,
			slave.Port,
			slave.User,
			slave.Password,
			slave.DBName,
		)
		slaveDSNs = append(slaveDSNs, dsn)
	}

	opts := wbdb.Options{
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		MaxOpenConns:    cfg.DB.MaxOpenConns,
	}

	db, err := wbdb.New(masterDSN, slaveDSNs, &opts)
	if err != nil {
		wbzlog.Logger.Debug().Msg("failed to connect to Postgres")
		return nil, err
	}

	wbzlog.Logger.Info().Msg("connected to Postgres")

	queryTimeout := cfg.DB.QueryTimeout

	return &Repository{db: db, retry: &cfg.Retry, queryTimeout: queryTimeout}, nil
}

// Close closes all database connections.
func (r *Repository) Close() error {
	if err := r.db.Master.Close(); err != nil {
		wbzlog.Logger.Debug().Msg("failed to close Postgres master connection")
		return err
	}

	for _, slave := range r.db.Slaves {
		if slave != nil {
			if err := slave.Close(); err != nil {
				wbzlog.Logger.Debug().Msg("failed to close Postgres slave connection")
				return err
			}
		}
	}

	return nil
}
