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
		incomingReqId string
	}
	type want struct {
		newReqId bool
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
				newReqId: true,
			},
		},
		{
			name: "generate new request id",
			args: args{
				generate:      true,
				incomingReqId: "qweqwe",
			},
			want: want{
				newReqId: true,
			},
		},
		{
			name: "keep request id",
			args: args{
				generate:      false,
				incomingReqId: "qweqwe",
			},
			want: want{
				newReqId: false,
			},
		},
		{
			name: "empty request id",
			args: args{
				generate:      false,
				incomingReqId: "",
			},
			want: want{
				newReqId: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := New(&Config{Generate: tt.args.generate})

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.args.incomingReqId != "" {
				r.Header.Set("X-Request-ID", tt.args.incomingReqId)
			}
			s := &stub{}
			mw.Chain(s)
			mw.ServeHTTP(nil, r)

			reqId := s.r.Header.Get("X-Request-ID")

			require.True(t, s.wasCalled)

			if tt.want.newReqId {
				assert.NotEmpty(t, reqId)
				assert.NotEqual(t, tt.args.incomingReqId, reqId)
			} else {
				assert.Equal(t, tt.args.incomingReqId, reqId)
			}
		})
	}
}
