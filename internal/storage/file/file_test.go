package file

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/adwski/shorty/internal/storage/memory"
	"github.com/adwski/shorty/internal/storage/memory/db"

	"github.com/adwski/shorty/internal/storage"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFileStore(t *testing.T) {
	type args struct {
		key string
		url string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "store and get",
			args: args{
				key: "aaa",
				url: "https://bbb.ccc",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeFile := "/tmp/testFile" // can we actually mock this?
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			fs, err := New(&Config{
				FilePath:               storeFile,
				IgnoreContentOnStartup: true, // with mock we don't have to do this :(
				Logger:                 zap.NewExample(),
			})
			require.NoError(t, err)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go fs.Run(ctx, wg)

			// store
			err = fs.Store(tt.args.key, tt.args.url, false)
			require.NoError(t, err)

			// get
			var url string
			url, err = fs.Get(tt.args.key)
			require.NoError(t, err)
			assert.Equal(t, tt.args.url, url)

			// stop persistence
			done := make(chan struct{})
			go func() {
				wg.Wait()
				done <- struct{}{}
			}()
			cancel()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("file storage did not shutdown in time")
			}

			// check persistence
			var (
				content []byte
				urlRec  db.URLRecord
			)
			content, err = os.ReadFile(storeFile)
			require.NoError(t, err)
			err = json.Unmarshal(content, &urlRec)
			require.NoError(t, err)
			_, err = uuid.FromString(urlRec.UUID)
			require.NoError(t, err)
			assert.Equal(t, tt.args.key, urlRec.ShortURL)
			assert.Equal(t, tt.args.url, urlRec.OriginalURL)
		})
	}
}

func TestStore_Get(t *testing.T) {
	type args struct {
		db  map[string]string
		key string
		url string
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "get existing",
			args: args{
				db:  map[string]string{"aaa": "https://bbb.ccc"},
				key: "aaa",
				url: "https://bbb.ccc",
			},
		},
		{
			name: "get not existing",
			args: args{
				db:  map[string]string{},
				key: "aaa",
				url: "https://bbb.ccc",
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
				fs.Memory.DB[k] = db.URLRecord{
					UUID:        uuid.Must(uuid.NewV4()).String(),
					ShortURL:    k,
					OriginalURL: v,
				}
			}

			url, err := fs.Get(tt.args.key)
			if tt.err == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.args.url, url)
			} else {
				assert.Equal(t, tt.err, err)
			}
		})
	}
}

func TestStore_Store(t *testing.T) {
	type args struct {
		key       string
		url       string
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
				key:       "aaa",
				url:       "https://ddd.eee",
				beforeDB:  map[string]string{"aaa": "https://bbb.ccc"},
				wantDB:    map[string]string{"aaa": "https://ddd.eee"},
				overwrite: true,
			},
		},
		{
			name: "store existing without overwrite",
			args: args{
				key:       "aaa",
				url:       "https://ddd.eee",
				beforeDB:  map[string]string{"aaa": "https://bbb.ccc"},
				wantDB:    map[string]string{"aaa": "https://bbb.ccc"},
				wantErr:   storage.ErrAlreadyExists,
				overwrite: false,
			},
		},
		{
			name: "store not existing",
			args: args{
				key:       "aaa",
				url:       "https://bbb.ccc",
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
				fs.Memory.DB[k] = db.URLRecord{
					UUID:        uuid.Must(uuid.NewV4()).String(),
					ShortURL:    k,
					OriginalURL: v,
				}
			}

			err := fs.Store(tt.args.key, tt.args.url, tt.args.overwrite)

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
