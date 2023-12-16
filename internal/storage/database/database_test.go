//go:build integration
// +build integration

package database

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/adwski/shorty/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	externalDBDSN = "postgres://shorty:shorty@localhost:5432/shorty?sslmode=disable"
)

var (
	db *Database
)

func TestMain(m *testing.M) {
	var (
		code int
	)
	log, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("logger init error", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		if db != nil {
			db.Close()
		}
	}()
	if db, err = New(ctx, &Config{
		Logger:  log,
		DSN:     externalDBDSN,
		Migrate: true,
		Trace:   false,
	}); err != nil {
		log.Error("error connecting to db", zap.Error(err), zap.String("dsn", externalDBDSN))
		code = 1
	} else {
		code = m.Run()
	}
	defer os.Exit(code)
}

func TestDatabase_Get(t *testing.T) {
	type args struct {
		urlInDB           *storage.URL
		get               string
		cleanupTestHashes bool
	}
	type want struct {
		err error
		url string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "simple get",
			args: args{
				urlInDB: &storage.URL{
					Short: "test123",
					Orig:  "http://test123.test123/test123",
				},
				get:               "test123",
				cleanupTestHashes: true,
			},
			want: want{
				url: "http://test123.test123/test123",
			},
		},
		{
			name: "get not existing",
			args: args{
				get: "test123",
			},
			want: want{
				err: storage.ErrNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			// prepare data
			if tt.args.urlInDB != nil {
				tag, errT := db.pool.Exec(ctx, "insert into urls (hash, orig) values ($1, $2)",
					tt.args.urlInDB.Short, tt.args.urlInDB.Orig)
				require.NoError(t, errT)
				require.Equal(t, int64(1), tag.RowsAffected())
			}
			// test
			url, errU := db.Get(ctx, tt.args.get)
			require.Equal(t, tt.want.err, errU)
			if tt.want.err == nil {
				assert.Equal(t, tt.want.url, url)
			}

			// clean up
			if tt.args.cleanupTestHashes {
				cleanUpTestHashes(ctx, t, db.pool)
			}
		})
	}
}

func TestDatabase_Store(t *testing.T) {
	type args struct {
		urlInDB           *storage.URL
		storeURL          storage.URL
		overwrite         bool
		cleanupTestHashes bool
	}
	type want struct {
		err  error
		hash string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "simple store",
			args: args{
				storeURL: storage.URL{
					Short: "test456",
					Orig:  "http://test456.test456/test456",
				},
				cleanupTestHashes: true,
			},
		},
		{
			name: "store same overwrite",
			args: args{
				urlInDB: &storage.URL{
					Short: "test456",
					Orig:  "http://test456.test456/test456",
				},
				storeURL: storage.URL{
					Short: "test456",
					Orig:  "http://test456.test456/test456",
				},
				overwrite:         true,
				cleanupTestHashes: true,
			},
		},
		{
			name: "store same no overwrite",
			args: args{
				urlInDB: &storage.URL{
					Short: "test456",
					Orig:  "http://test456.test456/test456",
				},
				storeURL: storage.URL{
					Short: "test456",
					Orig:  "http://test789.test789/test789",
				},
				cleanupTestHashes: true,
			},
			want: want{
				err: storage.ErrAlreadyExists,
			},
		},
		{
			name: "store same orig",
			args: args{
				urlInDB: &storage.URL{
					Short: "test789",
					Orig:  "http://test789.test789/test789",
				},
				storeURL: storage.URL{
					Short: "test456",
					Orig:  "http://test789.test789/test789",
				},
				cleanupTestHashes: true,
			},
			want: want{
				err:  storage.ErrConflict,
				hash: "test789",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			// prepare data
			if tt.args.urlInDB != nil {
				tag, errT := db.pool.Exec(ctx, "insert into urls (hash, orig) values ($1, $2)",
					tt.args.urlInDB.Short, tt.args.urlInDB.Orig)
				require.NoError(t, errT)
				require.Equal(t, int64(1), tag.RowsAffected())
			}

			// test
			hash, errU := db.Store(ctx, &tt.args.storeURL, tt.args.overwrite)
			require.Equal(t, tt.want.err, errU)
			assert.Equal(t, tt.want.hash, hash)

			// clean up
			if tt.args.cleanupTestHashes {
				cleanUpTestHashes(ctx, t, db.pool)
			}
		})
	}
}

func TestDatabase_StoreBatch(t *testing.T) {
	type args struct {
		urlInDB           *storage.URL
		batch             []storage.URL
		overwrite         bool
		cleanupTestHashes bool
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "simple store",
			args: args{
				batch: []storage.URL{
					{
						Short: "test456",
						Orig:  "http://test456.test456/test456",
					},
					{
						Short: "test789",
						Orig:  "http://test789.test789/test789",
					},
				},
				cleanupTestHashes: true,
			},
		},
		{
			name: "store existing",
			args: args{
				urlInDB: &storage.URL{
					Short: "test456",
					Orig:  "http://test456.test456/test456",
				},
				batch: []storage.URL{
					{
						Short: "test456",
						Orig:  "http://test456.test456/test456",
					},
					{
						Short: "test789",
						Orig:  "http://test789.test789/test789",
					},
				},
				cleanupTestHashes: true,
			},
			want: want{
				err: storage.ErrAlreadyExists,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			// prepare data
			if tt.args.urlInDB != nil {
				tag, errT := db.pool.Exec(ctx, "insert into urls (hash, orig) values ($1, $2)",
					tt.args.urlInDB.Short, tt.args.urlInDB.Orig)
				require.NoError(t, errT)
				require.Equal(t, int64(1), tag.RowsAffected())
			}

			// test
			err := db.StoreBatch(ctx, tt.args.batch)
			require.Equal(t, tt.want.err, err)

			// clean up
			if tt.args.cleanupTestHashes {
				cleanUpTestHashes(ctx, t, db.pool)
			}
		})
	}
}

func cleanUpTestHashes(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	tag, errE := pool.Exec(ctx, "delete from urls where hash like 'test%'")
	if errE != nil {
		require.NoError(t, errE)
	} else {
		t.Log("cleanup done, affected rows", tag.RowsAffected())
	}
}
