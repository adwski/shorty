//go:build integration
// +build integration

package database

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

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
		cleanUP(ctx, log, db.pool) // clean up here in case some test failed
		code = m.Run()
	}
	defer os.Exit(code)
}

func cleanUP(ctx context.Context, log *zap.Logger, pool *pgxpool.Pool) {
	tag, err := pool.Exec(ctx, "delete from urls where hash like 'test%'")
	if err != nil {
		panic(err)
	}
	log.Debug("cleaned up before tests", zap.Int64("affected", tag.RowsAffected()))
}

func TestDatabase_Get(t *testing.T) {
	type args struct {
		urlInDB *storage.URL
		get     string
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
				get: "test123",
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
			cleanUpTestHashes(ctx, t, db.pool)
		})
	}
}

func TestDatabase_Store(t *testing.T) {
	type args struct {
		urlInDB   *storage.URL
		storeURL  storage.URL
		overwrite bool
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
					Short:  "test456",
					Orig:   "http://test456.test456/test456",
					UserID: "testuser",
				},
			},
		},
		{
			name: "store same overwrite",
			args: args{
				urlInDB: &storage.URL{
					Short:  "test456",
					Orig:   "http://test456.test456/test456",
					UserID: "testuser",
				},
				storeURL: storage.URL{
					Short:  "test456",
					Orig:   "http://test456.test456/test456",
					UserID: "testuser",
				},
				overwrite: true,
			},
		},
		{
			name: "store same no overwrite",
			args: args{
				urlInDB: &storage.URL{
					Short:  "test456",
					Orig:   "http://test456.test456/test456",
					UserID: "testuser",
				},
				storeURL: storage.URL{
					Short:  "test456",
					Orig:   "http://test789.test789/test789",
					UserID: "testuser",
				},
			},
			want: want{
				err: storage.ErrAlreadyExists,
			},
		},
		{
			name: "store same orig",
			args: args{
				urlInDB: &storage.URL{
					Short:  "test789",
					Orig:   "http://test789.test789/test789",
					UserID: "testuser",
				},
				storeURL: storage.URL{
					Short:  "test456",
					Orig:   "http://test789.test789/test789",
					UserID: "testuser",
				},
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
			cleanUpTestHashes(ctx, t, db.pool)
		})
	}
}

func TestDatabase_StoreBatch(t *testing.T) {
	type args struct {
		urlInDB *storage.URL
		batch   []storage.URL
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
						Short:  "test456",
						Orig:   "http://test456.test456/test456",
						UserID: "testuser",
					},
					{
						Short:  "test789",
						Orig:   "http://test789.test789/test789",
						UserID: "testuser",
					},
				},
			},
		},
		{
			name: "store existing",
			args: args{
				urlInDB: &storage.URL{
					Short:  "test456",
					Orig:   "http://test456.test456/test456",
					UserID: "testuser",
				},
				batch: []storage.URL{
					{
						Short:  "test456",
						Orig:   "http://test456.test456/test456",
						UserID: "testuser",
					},
					{
						Short:  "test789",
						Orig:   "http://test789.test789/test789",
						UserID: "testuser",
					},
				},
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
			cleanUpTestHashes(ctx, t, db.pool)
		})
	}
}

func TestDatabase_ListUserURLs(t *testing.T) {
	type args struct {
		urlsInDB []storage.URL
		userID   string
	}
	type want struct {
		err    error
		hashes []string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "list urls",
			args: args{
				urlsInDB: []storage.URL{
					{
						Short:  "test456",
						Orig:   "http://test456.test456/test456",
						UserID: "testuser",
					},
					{
						Short:  "test789",
						Orig:   "http://test789.test789/test789",
						UserID: "testuser2",
					},
				},
				userID: "testuser",
			},
			want: want{
				err:    nil,
				hashes: []string{"test456"},
			},
		},
		{
			name: "empty urls",
			args: args{
				urlsInDB: []storage.URL{
					{
						Short:  "test456",
						Orig:   "http://test456.test456/test456",
						UserID: "testuser",
					},
					{
						Short:  "test789",
						Orig:   "http://test789.test789/test789",
						UserID: "testuser2",
					},
				},
				userID: "testuser3",
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
			for _, u := range tt.args.urlsInDB {
				tag, errT := db.pool.Exec(ctx, "insert into urls (hash, orig, userid) values ($1, $2, $3)",
					u.Short, u.Orig, u.UserID)
				require.NoError(t, errT)
				require.Equal(t, int64(1), tag.RowsAffected())
			}

			// test
			urls, err := db.ListUserURLs(ctx, tt.args.userID)
			require.Equal(t, tt.want.err, err)

			if tt.want.err == nil {
				hashes := tt.want.hashes
				require.Equal(t, len(tt.want.hashes), len(urls))

				sort.Slice(urls, func(i, j int) bool { return urls[i].Short > urls[j].Short })
				sort.Slice(hashes, func(i, j int) bool { return hashes[i] > hashes[j] })

				for i, url := range urls {
					assert.Equal(t, hashes[i], url.Short)
				}
			}

			// clean up
			cleanUpTestHashes(ctx, t, db.pool)
		})
	}
}

func TestDatabase_DeleteUserURLs(t *testing.T) {
	type args struct {
		urlsInDB        []storage.URL
		urlsForDeletion []storage.URL
	}
	type want struct {
		err        error
		affected   int64
		urlsRemain []storage.URL
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "delete url",
			args: args{
				urlsInDB: []storage.URL{
					{
						Short:  "test456",
						Orig:   "http://test456.test456/test456",
						UserID: "testuser",
					},
					{
						Short:  "test789",
						Orig:   "http://test789.test789/test789",
						UserID: "testuser2",
					},
				},
				urlsForDeletion: []storage.URL{
					{
						Short:  "test456",
						UserID: "testuser",
					},
				},
			},
			want: want{
				affected: 1,
				urlsRemain: []storage.URL{
					{
						Short:  "test789",
						Orig:   "http://test789.test789/test789",
						UserID: "testuser2",
					},
				},
			},
		},
		{
			name: "delete not existing urls",
			args: args{
				urlsInDB: []storage.URL{
					{
						Short:  "test456",
						Orig:   "http://test456.test456/test456",
						UserID: "testuser",
					},
					{
						Short:  "test789",
						Orig:   "http://test789.test789/test789",
						UserID: "testuser2",
					},
				},
				urlsForDeletion: []storage.URL{
					{
						Short:  "test4567",
						UserID: "testuser3",
					},
				},
			},
			want: want{
				affected: 0,
				urlsRemain: []storage.URL{
					{
						Short:  "test456",
						Orig:   "http://test456.test456/test456",
						UserID: "testuser",
					},
					{
						Short:  "test789",
						Orig:   "http://test789.test789/test789",
						UserID: "testuser2",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// prepare data
			for _, u := range tt.args.urlsInDB {
				tag, errT := db.pool.Exec(ctx, "insert into urls (hash, orig, userid) values ($1, $2, $3)",
					u.Short, u.Orig, u.UserID)
				require.NoError(t, errT)
				require.Equal(t, int64(1), tag.RowsAffected())
			}
			<-time.After(time.Second) // allow some delay to ensure timestamp differences on deletion
			// test
			affected, err := db.DeleteUserURLs(ctx, tt.args.urlsForDeletion)
			require.Equal(t, tt.want.err, err)
			require.Equal(t, tt.want.affected, affected)

			for _, url := range tt.want.urlsRemain {
				orig, errU := db.Get(ctx, url.Short)
				require.NoError(t, errU)
				assert.Equal(t, url.Orig, orig)
			}

			// clean up
			cleanUpTestHashes(ctx, t, db.pool)
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
