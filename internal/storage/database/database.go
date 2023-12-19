package database

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

const (
	urlsIndexHash = "urls_hash"
	urlsIndexOrig = "urls_orig_key"
)

// Database is a relational database storage type.
type Database struct {
	pool        *pgxpool.Pool
	config      *pgxpool.Config
	log         *zap.Logger
	dsn         string
	doMigration bool
}

func (db *Database) Close() {
	db.log.Debug("closing pgx connection pool")
	db.pool.Close()
	db.log.Debug("pgx connection pool is closed")
}

func (db *Database) Ping(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("db ping unsuccessful: %w", err)
	}
	return nil
}

func (db *Database) Store(ctx context.Context, url *storage.URL, overwrite bool) (string, error) {
	if overwrite {
		// we could not do it in one query here
		// because of conflict with unique orig constraint
		return db.storeWithOverwrite(ctx, url)
	}

	query := `insert into urls(hash, orig, uid) values ($1,$2, $3)`
	tag, err := db.pool.Exec(ctx, query, url.Short, url.Orig, url.UID)

	if err == nil {
		if tag.RowsAffected() != 1 {
			return "", fmt.Errorf("affected rows: %d, expected: 1", tag.RowsAffected())
		}
		return "", nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return "", fmt.Errorf("unknown postgres error: %w", err)
	}
	if pgErr.Code == pgerrcode.UniqueViolation {
		if pgErr.ConstraintName == urlsIndexHash {
			return "", storage.ErrAlreadyExists
		}
		if pgErr.ConstraintName == urlsIndexOrig {
			storedHash, errGet := db.getHashByURL(ctx, url.Orig)
			if errGet != nil {
				return "", errGet
			}
			return storedHash, storage.ErrConflict
		}
	}
	return "", fmt.Errorf("postgres error: %w", err)
}

func (db *Database) storeWithOverwrite(ctx context.Context, url *storage.URL) (string, error) {
	if storedURL, err := db.Get(ctx, url.Short); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			// no records, call store with no overwrite
			return db.Store(ctx, url, false)
		}
		// some other error
		return "", err
	} else if storedURL == url.Orig {
		// stored orig value is the same, no need to update
		return "", nil
	}
	// record exists, update it
	return "", db.updateOrig(ctx, url)
}

func (db *Database) updateOrig(ctx context.Context, url *storage.URL) error {
	query := "update urls set hash = $1 where orig = $2"
	tag, err := db.pool.Exec(ctx, query, url.Short, url.Orig)
	if err != nil {
		return fmt.Errorf("database update error: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("update affected rows: %d, expected: 1", tag.RowsAffected())
	}
	return nil
}

func (db *Database) StoreBatch(ctx context.Context, urls []storage.URL) error {
	batch := &pgx.Batch{} // implicit BEGIN and COMMIT
	for _, url := range urls {
		// There's an upper limit for number of queries that can be bundled in single batch,
		// but it depends on a particular setup.
		// https://youtu.be/sXMSWhcHCf8?t=33m55s
		batch.Queue(`insert into urls(hash, orig, uid) values ($1, $2, $3)`, url.Short, url.Orig, url.UID)
	}

	if err := db.pool.SendBatch(ctx, batch).Close(); err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			return fmt.Errorf("unknown postgres error: %w", err)
		}
		if pgErr.Code == pgerrcode.UniqueViolation {
			return storage.ErrAlreadyExists
		}
		return fmt.Errorf("pgx batch error: %w", err)
	}
	return nil
}

func (db *Database) Get(ctx context.Context, hash string) (url string, err error) {
	err = db.pool.QueryRow(ctx, `select orig from urls where hash = $1`, hash).Scan(&url)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		err = storage.ErrNotFound
	}
	return
}

func (db *Database) ListUserURLs(ctx context.Context, uid string) ([]*storage.URL, error) {
	query := `select hash, orig from urls where uid = $1`
	rows, err := db.pool.Query(ctx, query, uid)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		err = storage.ErrNotFound
		return nil, err
	}
	db.log.Debug("executing query",
		zap.String("uid", uid),
		zap.Int64("affected rows", rows.CommandTag().RowsAffected()))
	// Use generic CollectRows()
	// https://youtu.be/sXMSWhcHCf8?t=995
	urls, errR := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*storage.URL, error) {
		db.log.Debug("scanning one row")
		var url storage.URL
		if errS := row.Scan(&url.Short, &url.Orig); errS != nil {
			return nil, fmt.Errorf("error while scanning row: %w", errS)
		}
		return &url, nil
	})
	if errR != nil {
		return nil, fmt.Errorf("error while collecting rows: %w", errR)
	}
	return urls, nil
}

func (db *Database) getHashByURL(ctx context.Context, url string) (hash string, err error) {
	err = db.pool.QueryRow(ctx, `select hash from urls where orig = $1`, url).Scan(&hash)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		err = storage.ErrNotFound
	}
	return
}
