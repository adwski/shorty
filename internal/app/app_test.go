package app

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestNewShorty(t *testing.T) {

	cfg := &ShortyConfig{
		Host:         "xxx.yyy",
		ServedScheme: "http",
	}
	shorty := NewShorty(cfg)

	urls := []string{
		"http://aaa.bbb",
		"https://ccc.ddd/123",
		"https://eee.fff:4567/890",
	}
	for _, url_ := range urls {
		t.Run("Storing and getting "+url_, func(t *testing.T) {

			//-----------------------------------------------------
			// Store URL
			//-----------------------------------------------------
			body := []byte(url_)
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			r.Header.Set("Content-Type", "text/plain")
			r.Header.Set("Content-Length", strconv.Itoa(len(body)))

			w := httptest.NewRecorder()

			// Execute
			shorty.ServeHTTP(w, r)
			res := w.Result()

			// Check status
			assert.Equal(t, http.StatusCreated, res.StatusCode)

			// Check headers
			assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))

			// Check body
			defer res.Body.Close()
			urlBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			u, errU := url.Parse(string(urlBody))
			require.NoError(t, errU)
			assert.Equal(t, cfg.Host, u.Host)
			assert.Equal(t, cfg.ServedScheme, u.Scheme)

			//-----------------------------------------------------
			// Get redirect
			//-----------------------------------------------------

			r = httptest.NewRequest(http.MethodGet, u.Path, nil)

			w = httptest.NewRecorder()

			// Execute
			shorty.ServeHTTP(w, r)
			res = w.Result()
			defer res.Body.Close()

			// Check status
			assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)

			// Check headers
			assert.Equal(t, url_, res.Header.Get("Location"))
		})
	}
}
