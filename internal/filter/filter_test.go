package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFilter(t *testing.T) {
	type args struct {
		xFF          string
		xRealIP      string
		remoteAdd    string
		trusted      string
		trustXFF     bool
		trustXRealIP bool
	}
	type want struct {
		match  bool
		cfgErr bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "trusted remote addr",
			args: args{
				remoteAdd: "127.0.0.1:1111",
				trusted:   "127.0.0.0/8",
			},
			want: want{match: true},
		},
		{
			name: "trusted xff",
			args: args{
				trustXFF:  true,
				xFF:       "10.10.10.10",
				remoteAdd: "127.0.0.1:1111",
				trusted:   "10.10.10.0/24",
			},
			want: want{match: true},
		},
		{
			name: "trusted x-real-ip",
			args: args{
				trustXRealIP: true,
				xRealIP:      "10.10.10.10",
				remoteAdd:    "127.0.0.1:1111",
				trusted:      "10.10.10.0/24",
			},
			want: want{match: true},
		},
		{
			name: "trusted xff, untrusted x-real-ip",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xRealIP:      "20.10.10.10",
				xFF:          "10.10.10.10",
				trusted:      "10.10.10.0/24",
			},
			want: want{match: true},
		},
		{
			name: "trusted x-real-ip, untrusted xff",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xRealIP:      "20.10.10.10",
				xFF:          "10.10.10.10",
				trusted:      "20.10.10.0/24",
			},
			want: want{match: true},
		},
		{
			name: "trusted x-real-ip, malformed xff",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xRealIP:      "20.10.10.10",
				xFF:          "10.10.10.10, asdasd",
				trusted:      "20.10.10.0/24",
			},
			want: want{match: true},
		},
		{
			name: "malformed xff",
			args: args{
				trustXFF: true,
				xFF:      "10.10.10.10, asdasd",
				trusted:  "10.10.10.0/24",
			},
			want: want{match: false},
		},
		{
			name: "empty x-real-ip",
			args: args{
				trustXRealIP: true,
				xRealIP:      "",
				trusted:      "10.10.10.0/24",
			},
			want: want{match: false},
		},
		{
			name: "empty xff",
			args: args{
				trustXFF: true,
				xFF:      "",
				trusted:  "10.10.10.0/24",
			},
			want: want{match: false},
		},
		{
			name: "trust no one",
			args: args{
				trusted: "",
			},
			want: want{match: false},
		},
		{
			name: "no match",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xRealIP:      "4.4.4.4",
				xFF:          "5.5.5.5, 6.6.6.6",
				remoteAdd:    "127.0.0.1:1111",
				trusted:      "1.1.1.0/24,2.2.2.0/24,3.3.3.0/24",
			},
			want: want{match: false},
		},
		{
			name: "cfg err",
			args: args{
				trusted: "1.1.1.1/1, 2.2.2.2",
			},
			want: want{cfgErr: true},
		},
		{
			name: "ipv6 x-real-ip match",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xRealIP:      "2000::1",
				xFF:          "5.5.5.5, 6.6.6.6",
				remoteAdd:    "127.0.0.1:1111",
				trusted:      "2000::/16",
			},
			want: want{match: true},
		},
		{
			name: "ipv6 xff no-match",
			args: args{
				trustXFF:  true,
				xFF:       "2000::1, 6.6.6.6",
				remoteAdd: "127.0.0.1:1111",
				trusted:   "2000::/16",
			},
			want: want{match: false},
		},
		{
			name: "ipv6 xff match",
			args: args{
				trustXFF:     true,
				trustXRealIP: true,
				xFF:          "2000::1, ff00::1",
				remoteAdd:    "[::1]:1111",
				trusted:      "2000::/16,ff00::/16",
			},
			want: want{match: true},
		},
		{
			name: "ipv6 remote addr match",
			args: args{
				remoteAdd: "[abcd::abcd]:1111",
				trusted:   "abcd::/16",
			},
			want: want{match: true},
		},
		{
			name: "ipv6 remote addr no-match",
			args: args{
				remoteAdd: "[abcd::abcd]:1111",
				trusted:   "abce::/16",
			},
			want: want{match: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, errL := zap.NewDevelopment()
			require.NoError(t, errL)

			// create filter
			f, err := New(&Config{
				Logger:             logger,
				Subnets:            tt.args.trusted,
				TrustXForwardedFor: tt.args.trustXFF,
				TrustXRealIP:       tt.args.trustXRealIP,
			})

			// check for config error
			if tt.want.cfgErr {
				assert.Error(t, err)
				assert.Nil(t, f)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, f)

			// call filter with params
			ok := f.CheckRequestParams(tt.args.remoteAdd, tt.args.xRealIP, tt.args.xFF)
			assert.Equal(t, tt.want.match, ok)
		})
	}
}
