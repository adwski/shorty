package shortener

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/adwski/shorty/internal/app/mockapp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_ShortenBatch(t *testing.T) {
	type args struct {
		batch       []BatchURL
		serveHost   string
		serveScheme string
		pathLen     uint
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "shorten batch",
			args: args{
				serveHost:   "aaa",
				serveScheme: "http",
				pathLen:     7,
				batch: []BatchURL{
					{
						ID:  "123",
						URL: "http://qwe.qwe",
					},
					{
						ID:  "456",
						URL: "http://asd.asd",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare storage
			st := mockapp.NewStorage(t)

			// Prepare mock calls
			st.EXPECT().StoreBatch(mock.Anything, mock.Anything).Once().Return(nil)

			// Init service
			svc := New(&Config{
				Store:        st,
				Logger:       zap.NewExample(),
				ServedScheme: tt.args.serveScheme,
				Host:         tt.args.serveHost,
				PathLength:   tt.args.pathLen,
			})

			// Prepare batch request
			body, err := json.Marshal(tt.args.batch)
			require.NoError(t, err)
			r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
			r.Header.Set("Content-Type", "application/json")

			// Do request
			w := httptest.NewRecorder()
			svc.ShortenBatch(w, r)
			res := w.Result()

			// Check basic response params
			require.Equal(t, http.StatusCreated, res.StatusCode)
			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

			// Check body
			resBody, errB := io.ReadAll(res.Body)
			require.NoError(t, errB)
			require.NoError(t, res.Body.Close())

			// Check response correctness
			var resShortened []BatchShortened
			require.NoError(t, json.Unmarshal(resBody, &resShortened))
			require.Equal(t, len(tt.args.batch), len(resShortened))
			for i := range resShortened {
				assert.Equal(t, tt.args.batch[i].ID, resShortened[i].ID)
				u, errU := url.Parse(resShortened[i].Short)
				require.NoError(t, errU)
				assert.Equal(t, int(tt.args.pathLen), len(u.Path)-1)
				assert.Equal(t, tt.args.serveHost, u.Host)
				assert.Equal(t, tt.args.serveScheme, u.Scheme)
			}
		})
	}
}
