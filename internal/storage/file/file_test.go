package file

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/storage/memory"
	"github.com/adwski/shorty/internal/storage/memory/db"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFileStore(t *testing.T) {
	tests := []struct {
		name    string
		args    *storage.URL
		wantErr bool
	}{
		{
			name: "store and get",
			args: &storage.URL{
				Short: "aaa",
				Orig:  "https://bbb.ccc",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)
			ctx, cancel := context.WithCancel(context.Background())
			fStore, err := os.CreateTemp("", "shorty-test-db-*.")
			require.NoError(t, err)
			defer func() { _ = os.Remove(fStore.Name()) }()

			fs, err := New(ctx, &Config{
				FilePath: fStore.Name(),
				Logger:   logger,
			})
			require.NoError(t, err)

			// store
			_, err = fs.Store(ctx, tt.args, false)
			require.NoError(t, err)

			// get
			var url string
			url, err = fs.Get(ctx, tt.args.Short)
			require.NoError(t, err)
			assert.Equal(t, tt.args.Orig, url)

			// stop persistence
			cancel()
			fs.Close()

			// check persistence
			var (
				content []byte
				rec     db.Record
			)
			content, err = os.ReadFile(fStore.Name())
			require.NoError(t, err)
			require.NotEmpty(t, content)
			err = json.Unmarshal(content, &rec)
			require.NoError(t, err)
			_, err = uuid.FromString(rec.UUID)
			require.NoError(t, err)
			assert.Equal(t, tt.args.Short, rec.ShortURL)
			assert.Equal(t, tt.args.Orig, rec.OriginalURL)
		})
	}
}

func TestStore_Get(t *testing.T) {
	type args struct {
		db  map[string]string
		url *storage.URL
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "get existing",
			args: args{
				db: map[string]string{"aaa": "https://bbb.ccc"},
				url: &storage.URL{
					Short: "aaa",
					Orig:  "https://bbb.ccc",
				},
			},
		},
		{
			name: "get not existing",
			args: args{
				db: map[string]string{},
				url: &storage.URL{
					Short: "aaa",
					Orig:  "https://bbb.ccc",
				},
			},
			err: storage.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &File{
				Memory: memory.New(),
			}
			for k, v := range tt.args.db {
				fs.DB[k] = db.Record{
					UUID:        uuid.Must(uuid.NewV4()).String(),
					ShortURL:    k,
					OriginalURL: v,
				}
			}

			url, err := fs.Get(context.Background(), tt.args.url.Short)
			if tt.err == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.args.url.Orig, url)
			} else {
				assert.Equal(t, tt.err, err)
			}
		})
	}
}

func TestStore_Store(t *testing.T) {
	type args struct {
		url       *storage.URL
		beforeDB  map[string]string
		wantDB    map[string]string
		wantErr   error
		overwrite bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "store existing with overwrite",
			args: args{
				url: &storage.URL{
					Short: "aaa",
					Orig:  "https://ddd.eee",
				},
				beforeDB:  map[string]string{"aaa": "https://bbb.ccc"},
				wantDB:    map[string]string{"aaa": "https://ddd.eee"},
				overwrite: true,
			},
		},
		{
			name: "store existing without overwrite",
			args: args{
				url: &storage.URL{
					Short: "aaa",
					Orig:  "https://ddd.eee",
				},
				beforeDB:  map[string]string{"aaa": "https://bbb.ccc"},
				wantDB:    map[string]string{"aaa": "https://bbb.ccc"},
				wantErr:   storage.ErrAlreadyExists,
				overwrite: false,
			},
		},
		{
			name: "store not existing",
			args: args{
				url: &storage.URL{
					Short: "aaa",
					Orig:  "https://bbb.ccc",
				},
				beforeDB:  map[string]string{},
				wantDB:    map[string]string{"aaa": "https://bbb.ccc"},
				overwrite: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := File{
				Memory: memory.New(),
			}
			for k, v := range tt.args.beforeDB {
				fs.DB[k] = db.Record{
					UUID:        uuid.Must(uuid.NewV4()).String(),
					ShortURL:    k,
					OriginalURL: v,
				}
			}

			_, err := fs.Store(context.Background(), tt.args.url, tt.args.overwrite)

			if tt.args.wantErr != nil {
				require.NotNil(t, err)
				assert.ErrorIs(t, err, tt.args.wantErr)
			} else {
				require.Nil(t, err)
			}

			for k, v := range tt.args.wantDB {
				assert.Equal(t, v, fs.Memory.DB[k].OriginalURL)
			}
		})
	}
}
