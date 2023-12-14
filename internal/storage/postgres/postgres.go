package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/adwski/shorty/internal/storage"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Postgres is a PostgreSQL storage type.
type Postgres struct {
	pool        *pgxpool.Pool
	config      *pgxpool.Config
	log         *zap.Logger
	dsn         string
	doMigration bool
}

func (pg *Postgres) Init(ctx context.Context) error {
	if err := pg.migrate(); err != nil {
		return err
	}

	var err error
	pg.pool, err = pgxpool.NewWithConfig(ctx, pg.config)
	if err != nil {
		return fmt.Errorf("cannot create pgx connection pool: %w", err)
	}
	return nil
}

func (pg *Postgres) Close() {
	pg.log.Debug("closing pgx connection pool")
	pg.pool.Close()
	pg.log.Debug("pgx connection pool closed")
}

func (pg *Postgres) Ping(ctx context.Context) error {
	if err := pg.pool.Ping(ctx); err != nil {
		return fmt.Errorf("db ping unsuccessful: %w", err)
	}
	return nil
}

func (pg *Postgres) Store(ctx context.Context, hash, url string, overwrite bool) (string, error) {
	var (
		query = `insert into urls(hash, orig) values ($1,$2)`
	)
	if overwrite {
		query += ` ON CONFLICT (hash) DO UPDATE url = $2 where hash = $1`
	}
	tag, err := pg.pool.Exec(ctx, query, hash, url)
	if err == nil {
		if tag.RowsAffected() != 1 {
			return "", fmt.Errorf("affected rows: %d, expected: 1", tag.RowsAffected())
		}
		return "", nil
	}

	var pgErr *pgconn.PgError
	if !(errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation) {
		return "", fmt.Errorf("postgres error: %w", err)
	}
	storedHash, errGet := pg.getHashByURL(ctx, url)
	if errGet != nil {
		return "", errGet
	}
	return storedHash, storage.ErrConflict
}

func (pg *Postgres) StoreBatch(ctx context.Context, keys []string, urls []string) error {
	if len(keys) != len(urls) {
		return fmt.Errorf("incorrect number of arguments: keys: %d, urls: %d", len(keys), len(urls))
	}
	batch := &pgx.Batch{} // implicit BEGIN and COMMIT
	for i := range keys {
		// There's an upper limit for number of queries that can be bundled in single batch,
		// but it depends on a particular setup.
		// https://youtu.be/sXMSWhcHCf8?t=33m55s
		batch.Queue(`insert into urls(hash, orig) values ($1, $2)`, keys[i], urls[i])
	}
	if err := pg.pool.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("pgx batch error: %w", err)
	}
	return nil
}

func (pg *Postgres) Get(ctx context.Context, hash string) (url string, err error) {
	err = pg.pool.QueryRow(ctx, `select orig from urls where hash = $1`, hash).Scan(&url)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		err = storage.ErrNotFound
	}
	return
}

func (pg *Postgres) getHashByURL(ctx context.Context, url string) (hash string, err error) {
	err = pg.pool.QueryRow(ctx, `select hash from urls where orig = $1`, url).Scan(&hash)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		err = storage.ErrNotFound
	}
	return
}
