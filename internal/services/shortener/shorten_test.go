package shortener

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/adwski/shorty/internal/mocks/mockconfig"
	"github.com/stretchr/testify/mock"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Shorten(t *testing.T) {
	type args struct {
		pathLength     uint
		url            string
		headers        map[string]string
		addToStorage   map[string]string
		host           string
		servedScheme   string
		redirectScheme string
		json           bool
		invalid        bool
	}
	type want struct {
		status    int
		headers   map[string]string
		emptyBody bool
		url       string
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
				url:          "https://aaa.bbb",
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
			name: "store redirect json",
			args: args{
				pathLength:   10,
				url:          "https://aaa.bbb",
				servedScheme: "http",
				host:         "ccc.ddd",
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				json: true,
			},
			want: want{
				status: http.StatusCreated,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				url: "https://aaa.bbb",
			},
		},
		{
			name: "store same url",
			args: args{
				pathLength:   20,
				url:          "https://aaa.bbb",
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
				url:            "http://ccc.ddd",
				redirectScheme: "https",
				host:           "eee.fff",
				invalid:        true,
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
				url:          "https://aaa.bbb",
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
			st := mockconfig.NewStorage(t)

			if !tt.args.invalid {
				st.On("Store", mock.Anything, mock.Anything, false).Return(
					func(key, val string, _ bool) error {
						st.EXPECT().Get(key).Return(val, nil)
						return nil
					})
			}

			// Create Shortener
			svc := &Service{
				host:           tt.args.host,
				servedScheme:   tt.args.servedScheme,
				redirectScheme: tt.args.redirectScheme,
				pathLength:     tt.args.pathLength,
				store:          st,
				log:            zap.NewExample(),
			}

			var body []byte
			if tt.args.json {
				var jErr error
				body, jErr = json.Marshal(&ShortenRequest{URL: tt.args.url})
				require.NoError(t, jErr)
			} else {
				body = []byte(tt.args.url)
			}

			// Prepare request
			r := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body))
			for k, v := range tt.args.headers {
				r.Header.Set(k, v)
			}
			r.Header.Set("Content-Length", strconv.Itoa(len(body)))
			w := httptest.NewRecorder()

			// Execute
			if tt.args.json {
				svc.ShortenJSON(w, r)
			} else {
				svc.ShortenPlain(w, r)
			}
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
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())
			require.NotEqual(t, []byte{}, resBody)

			// Check return URL
			var u *url.URL
			if tt.args.json {
				var resp ShortenResponse
				err = json.Unmarshal(resBody, &resp)
				require.NoError(t, err)
				u, err = url.Parse(resp.Result)
				require.NoError(t, err)
			} else {
				u, err = url.Parse(string(resBody))
				require.NoError(t, err)
			}
			require.Equal(t, tt.args.pathLength, uint(len(u.Path)-1))
			require.Equal(t, u.Scheme, tt.args.servedScheme)
			require.Equal(t, u.Host, tt.args.host)

			// Check storage content
			storedURL, err3 := st.Get(u.Path[1:])
			require.NoError(t, err3)
			if tt.args.json {
				assert.Equal(t, tt.want.url, storedURL)
			} else {
				assert.Equal(t, tt.args.url, storedURL)
			}
		})
	}
}
