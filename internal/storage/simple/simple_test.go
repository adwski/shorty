package simple

import (
	"fmt"
	"github.com/adwski/shorty/internal/storage/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestNewSimple(t *testing.T) {
	type args struct {
		key string
		url string
	}
	tests := []struct {
		name    string
		args    args
		si      *Simple
		wantErr bool
	}{
		{
			name: "store and get",
			args: args{
				key: "aaa",
				url: "https://bbb.ccc",
			},
			si: NewSimple(&Config{PathLength: 10}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.si.Store(tt.args.key, tt.args.url)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			url, err2 := tt.si.Get(tt.args.key)
			require.NoError(t, err2)
			assert.Equal(t, tt.args.url, url)

		})
	}
}

func TestNewSimple_Get(t *testing.T) {
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
			err: errors.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			si := &Simple{
				st:  tt.args.db,
				mux: &sync.Mutex{},
			}

			url, err := si.Get(tt.args.key)
			if tt.err == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.args.url, url)
			} else {
				assert.Equal(t, tt.err, err)
			}

		})
	}
}

func TestNewSimple_Store(t *testing.T) {
	type args struct {
		key      string
		url      string
		beforeDB map[string]string
		wantDB   map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "store existing",
			args: args{
				key:      "aaa",
				url:      "https://ddd.eee",
				beforeDB: map[string]string{"aaa": "https://bbb.ccc"},
				wantDB:   map[string]string{"aaa": "https://ddd.eee"},
			},
		},
		{
			name: "store not existing",
			args: args{
				key:      "aaa",
				url:      "https://bbb.ccc",
				beforeDB: map[string]string{},
				wantDB:   map[string]string{"aaa": "https://bbb.ccc"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			si := &Simple{
				st:  tt.args.beforeDB,
				mux: &sync.Mutex{},
			}

			err := si.Store(tt.args.key, tt.args.url)

			require.NoError(t, err)
			assert.Equal(t, tt.args.wantDB, si.st)

		})
	}
}

func TestNewSimple_StoreUnique(t *testing.T) {
	type args struct {
		url        string
		runs       int
		pathLength uint
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "store",
			args: args{
				url:        "https://ddd.eee",
				runs:       1000,
				pathLength: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			si := NewSimple(&Config{PathLength: tt.args.pathLength})
			keys := make([]string, 0, tt.args.runs)
			for i := 0; i < tt.args.runs; i++ {
				key, err := si.StoreUnique(fmt.Sprintf("%s%d", tt.args.url, i))
				require.NoError(t, err)
				require.NotEmpty(t, key)
				require.Equal(t, tt.args.pathLength, uint(len(key)))
				for _, k := range keys {
					require.NotEqual(t, k, key)
				}
				keys = append(keys, key)
			}
			for i, k := range keys {
				url, err := si.Get(k)
				require.NoError(t, err, fmt.Sprintf("key is %s", k))
				assert.Equal(t, fmt.Sprintf("%s%d", tt.args.url, i), url)
			}
		})
	}
}
