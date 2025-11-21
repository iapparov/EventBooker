package postgres

import (
	"eventbooker/internal/config"
	"fmt"
	wbdb "github.com/wb-go/wbf/dbpg"
	wbzlog "github.com/wb-go/wbf/zlog"
)

type Postgres struct {
	db  *wbdb.DB
	cfg *config.RetrysConfig
}

func NewPostgres(cfg *config.AppConfig) (*Postgres, error) {
	masterDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBConfig.Master.Host,
		cfg.DBConfig.Master.Port,
		cfg.DBConfig.Master.User,
		cfg.DBConfig.Master.Password,
		cfg.DBConfig.Master.DBName,
	)

	slaveDSNs := make([]string, 0, len(cfg.DBConfig.Slaves))
	for _, slave := range cfg.DBConfig.Slaves {
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
	var opts wbdb.Options
	opts.ConnMaxLifetime = cfg.DBConfig.ConnMaxLifetime
	opts.MaxIdleConns = cfg.DBConfig.MaxIdleConns
	opts.MaxOpenConns = cfg.DBConfig.MaxOpenConns
	db, err := wbdb.New(masterDSN, slaveDSNs, &opts)
	if err != nil {
		wbzlog.Logger.Debug().Msg("Failed to connect to Postgres")
		return nil, err
	}
	wbzlog.Logger.Info().Msg("Connected to Postgres")
	return &Postgres{db: db, cfg: &cfg.RetrysConfig}, nil
}

func (p *Postgres) Close() error {
	err := p.db.Master.Close()
	if err != nil {
		wbzlog.Logger.Debug().Msg("Failed to close Postgres connection")
		return err
	}
	for _, slave := range p.db.Slaves {
		if slave != nil {
			err := slave.Close()
			if err != nil {
				wbzlog.Logger.Debug().Msg("Failed to close Postgres slave connection")
				return err
			}
		}
	}
	return nil
}
