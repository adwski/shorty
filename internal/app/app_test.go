package app

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/adwski/shorty/internal/mocks/mockconfig"
	"github.com/stretchr/testify/mock"

	"go.uber.org/zap"

	"github.com/adwski/shorty/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShorty(t *testing.T) {
	type args struct {
		url      string
		compress bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "http url",
			args: args{
				url: "http://aaa.bbb",
			},
		},
		{
			name: "https url with path",
			args: args{
				url: "https://ccc.ddd/123",
			},
		},
		{
			name: "https url compressed",
			args: args{
				url:      "https://ccc.ddd/123",
				compress: true,
			},
		},
		{
			name: "https url with port and path",
			args: args{
				url: "https://eee.fff:4567/890",
			},
		},
	}
	for _, tt := range tests {
		t.Run("Storing and getting "+tt.name, func(t *testing.T) {
			st := mockconfig.NewStorage(t)
			st.On("Store", mock.Anything, mock.Anything, false).Return(
				func(key, val string, _ bool) error {
					st.EXPECT().Get(key).Return(val, nil)
					return nil
				})

			cfg := &config.Shorty{
				Host:         "xxx.yyy",
				ServedScheme: "http",
				Logger:       zap.NewExample(),
				Storage:      st,
			}
			shorty := NewShorty(cfg)

			//-----------------------------------------------------
			// Store URL
			//-----------------------------------------------------
			var (
				bodyContent = tt.args.url
				body        = bytes.NewBuffer([]byte{})
				r           = httptest.NewRequest(http.MethodPost, "/", nil)
			)
			if tt.args.compress {
				gzw := gzip.NewWriter(body)
				_, err := gzw.Write([]byte(bodyContent))
				require.NoError(t, err)
				require.NoError(t, gzw.Close())
				r.Header.Set("Content-Encoding", "gzip")
			} else {
				body.Write([]byte(bodyContent))
			}
			r.Body = io.NopCloser(body)
			r.Header.Set("Content-Length", strconv.Itoa(body.Len()))
			r.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()

			// Execute
			shorty.server.Handler.ServeHTTP(w, r)
			res := w.Result()

			// Check status
			assert.Equal(t, http.StatusCreated, res.StatusCode)

			// Check headers
			assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))

			// Check body
			urlBody, err := io.ReadAll(res.Body)
			_ = res.Body.Close()
			require.NoError(t, err)
			u, errU := url.Parse(string(urlBody))
			require.NoError(t, errU)
			assert.Equal(t, cfg.Host, u.Host)
			assert.Equal(t, cfg.ServedScheme, u.Scheme)
			require.NotEmpty(t, u.Path)
			require.NotEqual(t, "/", u.Path)

			//-----------------------------------------------------
			// Get redirect
			//-----------------------------------------------------

			r = httptest.NewRequest(http.MethodGet, u.Path, nil)

			w = httptest.NewRecorder()

			// Execute
			shorty.server.Handler.ServeHTTP(w, r)
			res = w.Result()
			_ = res.Body.Close()

			// Check status
			assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)

			// Check headers
			assert.Equal(t, tt.args.url, res.Header.Get("Location"))
		})
	}
}
