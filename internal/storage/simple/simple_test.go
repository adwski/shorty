package simple

import (
	"sync"
	"testing"

	"github.com/adwski/shorty/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			si: NewSimple(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.si.Store(tt.args.key, tt.args.url, false)
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
				wantErr:   errors.ErrAlreadyExists,
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

			si := &Simple{
				st:  tt.args.beforeDB,
				mux: &sync.Mutex{},
			}

			err := si.Store(tt.args.key, tt.args.url, tt.args.overwrite)

			if tt.args.wantErr != nil {
				require.NotNil(t, err)
				assert.Equal(t, tt.args.wantErr.Error(), err.Error())
			} else {
				require.Nil(t, err)
			}
			assert.Equal(t, tt.args.wantDB, si.st)
		})
	}
}
