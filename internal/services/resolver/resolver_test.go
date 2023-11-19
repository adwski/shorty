package resolver

import (
	"github.com/adwski/shorty/internal/storage/simple"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Redirect(t *testing.T) {
	type args struct {
		pathLength   uint
		path         string
		addToStorage map[string]string
	}
	type want struct {
		status  int
		headers map[string]string
		body    string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "redirect existing",
			args: args{
				pathLength: 10,
				path:       "/qweasdzxcr",
				addToStorage: map[string]string{
					"qweasdzxcr": "https://aaa.bbb",
				},
			},
			want: want{
				status: http.StatusTemporaryRedirect,
				headers: map[string]string{
					"Location": "https://aaa.bbb",
				},
			},
		},
		{
			name: "redirect existing, different path length",
			args: args{
				pathLength: 20,
				path:       "/qweasdzxcr",
				addToStorage: map[string]string{
					"qweasdzxcr": "https://aaa.bbb",
				},
			},
			want: want{
				status: http.StatusTemporaryRedirect,
				headers: map[string]string{
					"Location": "https://aaa.bbb",
				},
			},
		},
		{
			name: "redirect not existing",
			args: args{
				pathLength: 10,
				path:       "/qweasd1xcr",
				addToStorage: map[string]string{
					"qweasdzxcr": "https://aaa.bbb",
				},
			},
			want: want{
				status: http.StatusNotFound,
				headers: map[string]string{
					"Location": "",
				},
			},
		},
		{
			name: "invalid request path",
			args: args{
				pathLength: 10,
				path:       "/qweasd&*zxcrаб",
				addToStorage: map[string]string{
					"qweasdzxcr123": "https://aaa.bbb",
				},
			},
			want: want{
				status: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				store: simple.New(),
				log:   zap.NewExample(),
			}
			for k, v := range tt.args.addToStorage {
				require.Nil(t, svc.store.Store(k, v, true))
			}

			r := httptest.NewRequest(http.MethodGet, tt.args.path, nil)
			w := httptest.NewRecorder()
			svc.Resolve(w, r)

			res := w.Result()

			assert.Equal(t, tt.want.status, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if tt.want.headers != nil {
				for k, v := range tt.want.headers {
					assert.Equal(t, v, res.Header.Get(k))
				}
			}
			assert.Equal(t, tt.want.body, string(resBody)) // JSONEq
		})
	}
}
