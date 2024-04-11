package status

import (
	"context"
	"errors"
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
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ping successful",
			args: args{pingErr: nil},
		},
		{
			name: "ping unsuccessful",
			args: args{pingErr: errors.New("test")},
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

			err = svc.Ping(ctx)
			if tt.args.pingErr != nil {
				assert.ErrorIs(t, err, tt.args.pingErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
