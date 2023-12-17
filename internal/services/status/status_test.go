package status

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adwski/shorty/internal/app/mockapp"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Status(t *testing.T) {
	type args struct {
		pingErr error
	}
	type want struct {
		status int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ping successful",
			args: args{pingErr: nil},
			want: want{status: http.StatusOK},
		},
		{
			name: "ping unsuccessful",
			args: args{pingErr: errors.New("test")},
			want: want{status: http.StatusInternalServerError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			st := mockapp.NewStorage(t)
			ctx := context.Background()

			st.EXPECT().Ping(ctx).Return(tt.args.pingErr)

			svc := &Service{
				store: st,
				log:   logger,
			}

			r := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()
			svc.Ping(w, r)
			res := w.Result()

			assert.Equal(t, tt.want.status, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())
			assert.Equal(t, "", string(resBody))
		})
	}
}
