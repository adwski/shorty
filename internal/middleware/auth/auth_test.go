package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adwski/shorty/internal/session"
	"github.com/adwski/shorty/internal/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stub struct {
	u         *user.User
	ok        bool
	wasCalled bool
}

func (s *stub) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	s.wasCalled = true
	s.u, s.ok = session.GetUserFromContext(r.Context())
}

func TestMiddleware(t *testing.T) {
	type args struct {
		user *user.User
	}
	type want struct {
		cookie bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "get auth cookie",
			args: args{user: nil},
			want: want{cookie: true},
		},
		{
			name: "auth with cookie",
			args: args{user: &user.User{ID: "bO64vpYQQpq3LlQRovQmlA"}},
			want: want{cookie: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)
			secret := "super-secret"
			mw := New(logger, secret)

			r := httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.args.user != nil {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"exp":    jwt.NewNumericDate(time.Now().Add(time.Hour)),
					"UserID": tt.args.user.ID,
				})
				signedToken, errS := token.SignedString([]byte(secret))
				require.NoError(t, errS)

				r.AddCookie(&http.Cookie{
					Name:  "shortySessID",
					Value: signedToken,
				})
			}

			s := &stub{}
			mw.HandleFunc(s)
			w := httptest.NewRecorder()
			mw.ServeHTTP(w, r)
			resp := w.Result()
			require.NoError(t, resp.Body.Close())

			require.True(t, s.wasCalled)
			require.True(t, s.ok)
			require.NotNil(t, s.u)

			if tt.want.cookie {
				require.Equal(t, 1, len(resp.Cookies()))
				tokenString := resp.Cookies()[0].Value
				require.NotEmpty(t, tokenString)
				token, errT := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
					return []byte(secret), nil
				})
				require.NoError(t, errT)
				assert.NotEmpty(t, token.Claims.(jwt.MapClaims)["UserID"])
			} else {
				require.Equal(t, 0, len(resp.Cookies()))
				assert.Equal(t, tt.args.user.ID, s.u.ID)
			}
		})
	}
}
