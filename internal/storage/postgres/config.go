package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgproto3"
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
	Logger  *zap.Logger
	DSN     string
	Migrate bool
	Trace   bool
}

func New(cfg *Config) (*Postgres, error) {
	if cfg.Logger == nil {
		return nil, errors.New("nil logger")
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

	var tracers map[uint32]*tracer
	if cfg.Trace {
		tracers = make(map[uint32]*tracer)
		pCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			pid := conn.PgConn().PID()
			tracers[pid] = newTracer(cfg.Logger, pid)
			conn.PgConn().Frontend().Trace(tracers[pid], pgproto3.TracerOptions{
				SuppressTimestamps: true,
				RegressMode:        true,
			})
			return nil
		}
	}

	return &Postgres{
		config:      pCfg,
		log:         cfg.Logger,
		dsn:         cfg.DSN,
		doMigration: cfg.Migrate,
		tracers:     tracers,
	}, nil
}

type tracer struct {
	log *zap.Logger
}

func newTracer(l *zap.Logger, pid uint32) *tracer {
	t := &tracer{
		log: l.With(zap.Uint32("pid", pid)),
	}
	t.log.Debug("spawning pgx tracer")
	return t
}

func (t *tracer) Write(b []byte) (int, error) {
	t.log.Debug("db trace",
		zap.String("trace", string(b)))
	return len(b), nil
}
