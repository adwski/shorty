package app

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/adwski/shorty/internal/app/mockapp"
	"github.com/adwski/shorty/internal/config"
	"github.com/adwski/shorty/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			st := mockapp.NewStorage(t)
			st.On("Store", mock.Anything, mock.Anything, false).Return(
				func(_ context.Context, url *model.URL, _ bool) (string, error) {
					t.Log("registering mock get", url)
					st.EXPECT().Get(mock.Anything, url.Short).Return(url.Orig, nil)
					return "", nil
				})

			cfg, err := config.New(logger)
			require.NoError(t, err)

			shorty, err := NewShorty(logger, st, cfg)
			require.NoError(t, err)

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
			shorty.http.Handler().ServeHTTP(w, r)
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
			assert.Equal(t, cfg.ServedHost, u.Host)
			assert.Equal(t, cfg.ServedScheme, u.Scheme)
			require.NotEmpty(t, u.Path)
			require.NotEqual(t, "/", u.Path)

			//-----------------------------------------------------
			// Get redirect
			//-----------------------------------------------------
			r = httptest.NewRequest(http.MethodGet, u.Path, nil)
			w = httptest.NewRecorder()

			// Execute
			shorty.http.Handler().ServeHTTP(w, r)
			res = w.Result()
			_ = res.Body.Close()

			// Check status
			assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)

			// Check headers
			assert.Equal(t, tt.args.url, res.Header.Get("Location"))
		})
	}
}
