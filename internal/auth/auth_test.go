package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adwski/shorty/internal/user"
	"github.com/stretchr/testify/require"
)

func TestAuth_CreateJWTCookie(t *testing.T) {
	type args struct {
		u         *user.User
		jwtSecret string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "create jwt cookie",
			args: args{
				u:         &user.User{ID: "tdGk2USqTvWW8jyz7HnhlA"},
				jwtSecret: "super-secret",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(tt.args.jwtSecret)
			cookie, err := a.CreateJWTCookie(tt.args.u)
			require.NoError(t, err)

			userFromCookie, errU := a.getUserFromJWT(cookie.Value)
			require.NoError(t, errU)
			require.Equal(t, tt.args.u.ID, userFromCookie.ID)
		})
	}
}

func TestAuth_GetUser(t *testing.T) {
	type args struct {
		u         *user.User
		jwtSecret string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "create jwt cookie",
			args: args{
				u:         &user.User{ID: "tdGk2USqTvWW8jyz7HnhlA"},
				jwtSecret: "super-secret",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(tt.args.jwtSecret)

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			token, err := a.newToken(tt.args.u)
			require.NoError(t, err)
			r.AddCookie(&http.Cookie{
				Name:  sessionCookieName,
				Value: token,
			})

			u, errU := a.GetUser(r)
			require.NoError(t, errU)

			require.Equal(t, tt.args.u.ID, u.ID)
		})
	}
}
