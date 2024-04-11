package filter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adwski/shorty/internal/filter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stub struct {
	called bool
}

func (s *stub) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	s.called = true
	rw.WriteHeader(http.StatusOK)
}

func TestMiddleware(t *testing.T) {
	type args struct {
		xFF          *string
		xRealIP      *string
		remoteAdd    *string
		trusted      string
		trustXFF     bool
		trustXRealIP bool
	}
	tests := []struct {
		name   string
		args   args
		called bool
		cfgErr bool
	}{
		{
			name: "trusted remote addr",
			args: args{
				remoteAdd: sPtr("127.0.0.1:1111"),
				trusted:   "127.0.0.0/8",
			},
			called: true,
		},
		{
			name: "trusted xff",
			args: args{
				trustXFF:  true,
				xFF:       sPtr("10.10.10.10"),
				remoteAdd: sPtr("127.0.0.1:1111"),
				trusted:   "10.10.10.0/24",
			},
			called: true,
		},
		{
			name: "trusted x-real-ip",
			args: args{
				trustXRealIP: true,
				xRealIP:      sPtr("10.10.10.10"),
				remoteAdd:    sPtr("127.0.0.1:1111"),
				trusted:      "10.10.10.0/24",
			},
			called: true,
		},
		{
			name: "trusted xff, untrusted x-real-ip",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xRealIP:      sPtr("20.10.10.10"),
				xFF:          sPtr("10.10.10.10"),
				trusted:      "10.10.10.0/24",
			},
			called: true,
		},
		{
			name: "trusted x-real-ip, untrusted xff",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xRealIP:      sPtr("20.10.10.10"),
				xFF:          sPtr("10.10.10.10"),
				trusted:      "20.10.10.0/24",
			},
			called: true,
		},
		{
			name: "trusted x-real-ip, malformed xff",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xRealIP:      sPtr("20.10.10.10"),
				xFF:          sPtr("10.10.10.10, asdasd"),
				trusted:      "20.10.10.0/24",
			},
			called: true,
		},
		{
			name: "malformed xff",
			args: args{
				trustXFF: true,
				xFF:      sPtr("10.10.10.10, asdasd"),
				trusted:  "10.10.10.0/24",
			},
			called: false,
		},
		{
			name: "empty x-real-ip",
			args: args{
				trustXRealIP: true,
				xRealIP:      sPtr(""),
				trusted:      "10.10.10.0/24",
			},
			called: false,
		},
		{
			name: "empty xff",
			args: args{
				trustXFF: true,
				xFF:      sPtr(""),
				trusted:  "10.10.10.0/24",
			},
			called: false,
		},
		{
			name: "block all",
			args: args{
				trusted: "",
			},
			called: false,
		},
		{
			name: "no match",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xRealIP:      sPtr("4.4.4.4"),
				xFF:          sPtr("5.5.5.5, 6.6.6.6"),
				remoteAdd:    sPtr("127.0.0.1:1111"),
				trusted:      "1.1.1.0/24,2.2.2.0/24,3.3.3.0/24",
			},
			called: false,
		},
		{
			name: "cfg err",
			args: args{
				trusted: "1.1.1.1/1, 2.2.2.2",
			},
			cfgErr: true,
		},
		{
			name: "ipv6 x-real-ip match",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xRealIP:      sPtr("2000::1"),
				xFF:          sPtr("5.5.5.5, 6.6.6.6"),
				remoteAdd:    sPtr("127.0.0.1:1111"),
				trusted:      "2000::/16",
			},
			called: true,
		},
		{
			name: "ipv6 xff no-match",
			args: args{
				trustXFF:  true,
				xFF:       sPtr("2000::1, 6.6.6.6"),
				remoteAdd: sPtr("127.0.0.1:1111"),
				trusted:   "2000::/16",
			},
			called: false,
		},
		{
			name: "ipv6 xff match",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xFF:          sPtr("2000::1, ff00::1"),
				remoteAdd:    sPtr("[::1]:1111"),
				trusted:      "2000::/16,ff00::/16",
			},
			called: true,
		},
		{
			name: "ipv6 remote addr match",
			args: args{
				remoteAdd: sPtr("[abcd::abcd]:1111"),
				trusted:   "abcd::/16",
			},
			called: true,
		},
		{
			name: "ipv6 remote addr no-match",
			args: args{
				remoteAdd: sPtr("[abcd::abcd]:1111"),
				trusted:   "abce::/16",
			},
			called: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, errL := zap.NewDevelopment()
			require.NoError(t, errL)

			// create middleware
			mw, err := New(&filter.Config{
				Logger:             logger,
				Subnets:            tt.args.trusted,
				TrustXForwardedFor: tt.args.trustXFF,
				TrustXRealIP:       tt.args.trustXRealIP,
			})

			// check for config error
			if tt.cfgErr {
				assert.Error(t, err)
				assert.Nil(t, mw)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, mw)

			// chain stub handler
			stubHandler := &stub{}
			mw.HandlerFunc(stubHandler)

			// create request
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.args.xFF != nil {
				r.Header.Set("X-Forwarded-For", *tt.args.xFF)
			}
			if tt.args.xRealIP != nil {
				r.Header.Set("X-Real-IP", *tt.args.xRealIP)
			}
			if tt.args.remoteAdd != nil {
				r.RemoteAddr = *tt.args.remoteAdd
			}

			// exec request
			w := httptest.NewRecorder()
			mw.ServeHTTP(w, r)
			resp := w.Result()
			_ = resp.Body.Close()

			// check response
			assert.Equal(t, tt.called, stubHandler.called)
			if tt.called {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			} else {
				assert.Equal(t, http.StatusForbidden, resp.StatusCode)
			}
		})
	}
}

func sPtr(s string) *string {
	return &s
}
