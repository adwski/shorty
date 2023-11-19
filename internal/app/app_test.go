package app

import (
	"bytes"
	"context"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/adwski/shorty/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShorty(t *testing.T) {
	cfg := &config.ShortyConfig{
		Host:         "xxx.yyy",
		ServedScheme: "http",
		Logger:       zap.NewExample(),
	}
	shorty, err := NewShorty(context.Background(), cfg)
	require.Nil(t, err)

	testURLs := []string{
		"http://aaa.bbb",
		"https://ccc.ddd/123",
		"https://eee.fff:4567/890",
	}
	for _, testURL := range testURLs {
		t.Run("Storing and getting "+testURL, func(t *testing.T) {

			//-----------------------------------------------------
			// Store URL
			//-----------------------------------------------------
			body := []byte(testURL)
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			r.Header.Set("Content-Type", "text/plain")
			r.Header.Set("Content-Length", strconv.Itoa(len(body)))

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
			assert.Equal(t, testURL, res.Header.Get("Location"))
		})
	}
}
