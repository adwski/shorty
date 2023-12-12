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
)

// Postgres is a PostgreSQL storage type.
type Postgres struct {
	*pgxpool.Pool
}

func (pg *Postgres) Store(ctx context.Context, hash, url string, overwrite bool) (string, error) {
	var (
		query = `insert into urls(hash, orig) values ($1,$2)`
	)
	if overwrite {
		query += ` ON CONFLICT (hash) DO UPDATE url = $2 where hash = $1`
	}
	tag, err := pg.Exec(ctx, query, hash, url)
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
	batch := &pgx.Batch{} // implicit BEGIN and COMMIT
	for i := range keys {
		// There's an upper limit for number of queries that can be bundled in single batch,
		// but it depends on a particular setup.
		// https://youtu.be/sXMSWhcHCf8?t=33m55s
		batch.Queue(`insert into urls(hash, orig) values ($1, $2)`, keys[i], urls[i])
	}
	err := pg.SendBatch(ctx, batch).Close()
	if err != nil {
		return fmt.Errorf("postgres batch error: %w", err)
	}
	return nil
}

func (pg *Postgres) Get(ctx context.Context, hash string) (url string, err error) {
	err = pg.QueryRow(ctx, `select orig from urls where hash = $1`, hash).Scan(&url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = storage.ErrNotFound
		}
	}
	return
}

func (pg *Postgres) getHashByURL(ctx context.Context, url string) (hash string, err error) {
	err = pg.QueryRow(ctx, `select hash from urls where orig = $1`, url).Scan(&hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = storage.ErrNotFound
		}
	}
	return
}
