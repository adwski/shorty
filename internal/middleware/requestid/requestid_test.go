package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adwski/shorty/internal/session"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stub struct {
	reqID     string
	wasCalled bool
	haveReqID bool
}

func (s *stub) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	s.wasCalled = true
	s.reqID, s.haveReqID = session.GetRequestID(r.Context())
}

func TestMiddleware(t *testing.T) {
	type args struct {
		trust         bool
		incomingReqID string
	}
	type want struct {
		newReqID bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "gen request id",
			args: args{
				trust: false,
			},
			want: want{
				newReqID: true,
			},
		},
		{
			name: "trust request id",
			args: args{
				trust:         true,
				incomingReqID: "asdasd",
			},
			want: want{
				newReqID: false,
			},
		},
		{
			name: "do not trust request id",
			args: args{
				trust:         false,
				incomingReqID: "qweqwe",
			},
			want: want{
				newReqID: true,
			},
		},
		{
			name: "empty request id",
			args: args{
				trust:         true,
				incomingReqID: "",
			},
			want: want{
				newReqID: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)
			mw := New(&Config{Trust: tt.args.trust, Logger: logger})

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.args.incomingReqID != "" {
				r.Header.Set("X-Request-ID", tt.args.incomingReqID)
			}
			s := &stub{}
			mw.HandlerFunc(s)
			mw.ServeHTTP(nil, r)
			require.True(t, s.wasCalled)
			require.True(t, s.haveReqID)

			if tt.want.newReqID {
				assert.NotEmpty(t, s.reqID)
				assert.NotEqual(t, tt.args.incomingReqID, s.reqID)
			} else {
				assert.Equal(t, tt.args.incomingReqID, s.reqID)
			}
		})
	}
}
