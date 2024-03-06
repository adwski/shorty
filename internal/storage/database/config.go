package database

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"

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

// Config holds Database configuration params.
type Config struct {
	Logger  *zap.Logger
	DSN     string
	Migrate bool
	Trace   bool
}

// New create new Database connector using Config.
func New(ctx context.Context, cfg *Config) (*Database, error) {
	if cfg.Logger == nil {
		return nil, errors.New("nil logger")
	}
	logger := cfg.Logger.With(zap.String("component", "database"))

	pCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DSN: %w", err)
	}
	preparePoolConfig(pCfg)
	if cfg.Trace {
		t := newTracers(logger)
		pCfg.AfterConnect = t.create()
		pCfg.BeforeClose = t.destroy()
	}

	db := &Database{
		config:      pCfg,
		log:         logger,
		dsn:         cfg.DSN,
		doMigration: cfg.Migrate,
	}

	if err = db.init(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

func preparePoolConfig(pCfg *pgxpool.Config) {
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
}

func (db *Database) init(ctx context.Context) error {
	if err := db.migrate(); err != nil {
		return err
	}

	var err error
	db.pool, err = pgxpool.NewWithConfig(ctx, db.config)
	if err != nil {
		return fmt.Errorf("cannot create pgx connection pool: %w", err)
	}
	return nil
}

type tracers struct {
	*sync.Map
	l *zap.Logger
}

func newTracers(logger *zap.Logger) *tracers {
	return &tracers{
		Map: &sync.Map{},
		l:   logger,
	}
}

func (t *tracers) destroy() func(conn *pgx.Conn) {
	return func(conn *pgx.Conn) {
		// Cleanup
		pid := conn.PgConn().PID()
		t.Delete(pid)
		t.l.Debug("pgx tracer destroyed", zap.Uint32("pid", pid))
	}
}

func (t *tracers) create() func(ctx context.Context, conn *pgx.Conn) error {
	return func(ctx context.Context, conn *pgx.Conn) error {
		// Spawn tracer
		pid := conn.PgConn().PID()
		t.l.Debug("spawning pgx tracer")
		tr := newTracer(t.l, pid)
		t.Store(pid, tr)
		conn.PgConn().Frontend().Trace(tr, pgproto3.TracerOptions{
			SuppressTimestamps: true,
			RegressMode:        true,
		})
		return nil
	}
}

type tracer struct {
	log *zap.Logger
}

func newTracer(l *zap.Logger, pid uint32) *tracer {
	return &tracer{
		log: l.With(zap.Uint32("pid", pid)),
	}
}

// Write writes trace string as one log message.
func (t *tracer) Write(b []byte) (int, error) {
	t.log.Debug("db trace",
		zap.String("trace", string(b)))
	return len(b), nil
}
