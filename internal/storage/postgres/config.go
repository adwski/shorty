package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	defaultConnectTimeout           = 3 * time.Second
	defaultConnectionIdle           = time.Minute
	defaultConnectionLifeTime       = time.Hour
	defaultConnectionLifeTimeJitter = 5 * time.Minute
	defaultMaxConns                 = 5
	defaultMinConns                 = 2
	defaultHealthCheckPeriod        = 3 * time.Second
)

type Config struct {
	Logger                       *zap.Logger
	DSN                          string
	Migrate                      bool
	EnforceDisableSSLOnMigration bool
}

func New(cfg *Config) (*Postgres, error) {
	if cfg.Logger == nil {
		return nil, errors.New("nil logger")
	}

	if cfg.Migrate {
		cfg.Logger.Warn("starting migration")
		if err := runMigrations(cfg.DSN, cfg.EnforceDisableSSLOnMigration); err != nil {
			return nil, fmt.Errorf("migration failure: %w", err)
		}
		cfg.Logger.Warn("migration is complete")
	}

	pCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DSN: %w", err)
	}

	pCfg.ConnConfig.Config.ConnectTimeout = defaultConnectTimeout

	// Choosing this mode because:
	// - Compatible with connection pollers
	// - Does not make two round trips
	// - Does not imply side effects after schema change
	// - We're using simple data types
	pCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	pCfg.MaxConnLifetime = defaultConnectionLifeTime
	pCfg.MaxConnLifetimeJitter = defaultConnectionLifeTimeJitter
	pCfg.MaxConnIdleTime = defaultConnectionIdle
	pCfg.MaxConns = defaultMaxConns
	pCfg.MinConns = defaultMinConns
	pCfg.HealthCheckPeriod = defaultHealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(context.Background(), pCfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create pgx connection pool: %w", err)
	}

	return &Postgres{Pool: pool}, nil
}
