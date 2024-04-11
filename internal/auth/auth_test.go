package auth

import (
	"testing"

	"github.com/adwski/shorty/internal/user"
	"github.com/stretchr/testify/require"
)

func TestAuth_CreateAndGetJWT(t *testing.T) {
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
			token, err := a.createJWT(tt.args.u)
			require.NoError(t, err)
			require.NotEmpty(t, token)

			userFromCookie, errU := a.getUserFromJWT(token)
			require.NoError(t, errU)
			require.Equal(t, tt.args.u.ID, userFromCookie.ID)
		})
	}
}
