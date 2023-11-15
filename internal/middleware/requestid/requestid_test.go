package requestid

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stub struct {
	r         *http.Request
	wasCalled bool
}

func (s *stub) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	s.wasCalled = true
	s.r = r
}

func TestMiddleware(t *testing.T) {
	type args struct {
		generate      bool
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
			name: "generate request id",
			args: args{
				generate: true,
			},
			want: want{
				newReqID: true,
			},
		},
		{
			name: "generate new request id",
			args: args{
				generate:      true,
				incomingReqID: "qweqwe",
			},
			want: want{
				newReqID: true,
			},
		},
		{
			name: "keep request id",
			args: args{
				generate:      false,
				incomingReqID: "qweqwe",
			},
			want: want{
				newReqID: false,
			},
		},
		{
			name: "empty request id",
			args: args{
				generate:      false,
				incomingReqID: "",
			},
			want: want{
				newReqID: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := New(&Config{Generate: tt.args.generate})

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.args.incomingReqID != "" {
				r.Header.Set("X-Request-ID", tt.args.incomingReqID)
			}
			s := &stub{}
			mw.Chain(s)
			mw.ServeHTTP(nil, r)

			reqID := s.r.Header.Get("X-Request-ID")

			require.True(t, s.wasCalled)

			if tt.want.newReqID {
				assert.NotEmpty(t, reqID)
				assert.NotEqual(t, tt.args.incomingReqID, reqID)
			} else {
				assert.Equal(t, tt.args.incomingReqID, reqID)
			}
		})
	}
}
