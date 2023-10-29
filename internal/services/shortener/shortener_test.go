package shortener

import (
	"bytes"
	"github.com/adwski/shorty/internal/storage/simple"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestService_Shorten(t *testing.T) {
	type args struct {
		pathLength     uint
		body           []byte
		addToStorage   map[string]string
		host           string
		servedScheme   string
		redirectScheme string
	}
	type want struct {
		status    int
		headers   map[string]string
		storage   map[string]string
		emptyBody bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "store redirect",
			args: args{
				pathLength:   10,
				body:         []byte("https://aaa.bbb"),
				servedScheme: "http",
				host:         "ccc.ddd",
			},
			want: want{
				status: http.StatusCreated,
				headers: map[string]string{
					"Content-Type": "text/plain",
				},
			},
		},
		{
			name: "store same url",
			args: args{
				pathLength:   20,
				body:         []byte("https://aaa.bbb"),
				servedScheme: "http",
				host:         "ccc.ddd",
				addToStorage: map[string]string{
					"qweqweqwe1": "https://aaa.bbb",
				},
			},
			want: want{
				status: http.StatusCreated,
				headers: map[string]string{
					"Content-Type": "text/plain",
				},
			},
		},
		{
			name: "store url wrong scheme",
			args: args{
				pathLength:     20,
				body:           []byte("http://ccc.ddd"),
				redirectScheme: "https",
				host:           "eee.fff",
			},
			want: want{
				status:    http.StatusBadRequest,
				emptyBody: true,
			},
		},
		{
			name: "store arbitrary scheme",
			args: args{
				pathLength:   20,
				body:         []byte("https://aaa.bbb"),
				servedScheme: "http",
				host:         "ccc.ddd",
			},
			want: want{
				status: http.StatusCreated,
				headers: map[string]string{
					"Content-Type": "text/plain",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Prepare storage
			simpleStore := simple.NewSimple(&simple.Config{PathLength: tt.args.pathLength})
			for k, v := range tt.args.addToStorage {
				_ = simpleStore.Store(k, v)
			}

			// Create Shortener
			svc := &Service{
				host:           tt.args.host,
				servedScheme:   tt.args.servedScheme,
				redirectScheme: tt.args.redirectScheme,
				store:          simpleStore,
			}

			// Prepare request
			r := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(tt.args.body))
			r.Header.Set("Content-Type", "text/plain")
			r.Header.Set("Content-Length", strconv.Itoa(len(tt.args.body)))
			w := httptest.NewRecorder()

			// Execute
			svc.Shorten(w, r)
			res := w.Result()

			// Check status code
			assert.Equal(t, tt.want.status, res.StatusCode)

			// Check headers
			if tt.want.headers != nil {
				for k, v := range tt.want.headers {
					assert.Equal(t, v, res.Header.Get(k))
				}
			}

			if tt.want.emptyBody {
				return
			}

			// Check body
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NotEqual(t, []byte{}, resBody)

			// Check return URL
			u, err2 := url.Parse(string(resBody))
			require.NoError(t, err2)
			require.Equal(t, tt.args.pathLength, uint(len(u.Path)-1))
			require.Equal(t, u.Scheme, tt.args.servedScheme)
			require.Equal(t, u.Host, tt.args.host)

			// Check storage content
			storage := simpleStore.DumpMap()
			require.Contains(t, storage, u.Path[1:])
			assert.Equal(t, storage[u.Path[1:]], string(tt.args.body))
		})
	}
}
