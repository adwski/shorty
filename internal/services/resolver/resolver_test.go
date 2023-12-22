package resolver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/adwski/shorty/internal/session"

	"github.com/adwski/shorty/internal/app/mockapp"

	"github.com/adwski/shorty/internal/storage"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Redirect(t *testing.T) {
	type args struct {
		pathLength   uint
		shortURL     string
		invalid      bool
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
				shortURL:   "qweasdzxcr",
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
				shortURL:   "qweasdzxcr",
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
				shortURL:   "qweasd1xcr",
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
				shortURL:   "qweasd&*zxcrаб",
				invalid:    true,
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
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			st := mockapp.NewStorage(t)

			if v, ok := tt.args.addToStorage[tt.args.shortURL]; !ok {
				if !tt.args.invalid {
					st.EXPECT().Get(mock.Anything, tt.args.shortURL).Return("", storage.ErrNotFound)
				}
			} else {
				st.EXPECT().Get(mock.Anything, tt.args.shortURL).Return(v, nil)
			}

			svc := &Service{
				store: st,
				log:   logger,
			}

			r := httptest.NewRequest(http.MethodGet, "/"+tt.args.shortURL, nil)
			r = r.WithContext(session.SetRequestID(r.Context(), "test-request"))
			w := httptest.NewRecorder()
			svc.Resolve(w, r)

			res := w.Result()

			assert.Equal(t, tt.want.status, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())

			if tt.want.headers != nil {
				for k, v := range tt.want.headers {
					assert.Equal(t, v, res.Header.Get(k))
				}
			}
			assert.Equal(t, tt.want.body, string(resBody)) // JSONEq
		})
	}
}
