package shortener

import (
	"context"
	"net/url"
	"testing"

	"github.com/adwski/shorty/internal/app/mockapp"
	"github.com/adwski/shorty/internal/model"
	"github.com/adwski/shorty/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_Shorten(t *testing.T) {
	type args struct {
		pathLength        uint
		url               string
		addToStorage      map[string]string
		host              string
		servedScheme      string
		redirectScheme    string
		doNotRegisterMock bool
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "store redirect",
			args: args{
				pathLength:   10,
				url:          "https://aaa.bbb",
				servedScheme: "http",
				host:         "ccc.ddd",
			},
		},
		{
			name: "store redirect json",
			args: args{
				pathLength:   10,
				url:          "https://aaa.bbb",
				servedScheme: "http",
				host:         "ccc.ddd",
			},
		},
		{
			name: "store same url",
			args: args{
				pathLength:   20,
				url:          "https://aaa.bbb",
				servedScheme: "http",
				host:         "ccc.ddd",
				addToStorage: map[string]string{
					"qweqweqwe1": "https://aaa.bbb",
				},
			},
		},
		{
			name: "store url wrong scheme",
			args: args{
				pathLength:        20,
				url:               "http://ccc.ddd",
				redirectScheme:    "https",
				host:              "eee.fff",
				doNotRegisterMock: true,
			},
			want: want{
				err: ErrUnsupportedURLScheme,
			},
		},
		{
			name: "store arbitrary scheme",
			args: args{
				pathLength:   20,
				url:          "https://aaa.bbb",
				servedScheme: "http",
				host:         "ccc.ddd",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			// Prepare storage
			st := mockapp.NewStorage(t)
			ctx := context.Background()

			// Register mock
			if !tt.args.doNotRegisterMock {
				st.On("Store", mock.Anything, mock.Anything, false).Return(
					func(_ context.Context, url *model.URL, _ bool) (string, error) {
						t.Log("registering mock get", url)
						st.EXPECT().Get(ctx, url.Short).Return(url.Orig, nil)
						return "", nil
					})
			}

			// Create Shortener
			svc := &Service{
				host:           tt.args.host,
				servedScheme:   tt.args.servedScheme,
				redirectScheme: tt.args.redirectScheme,
				pathLength:     tt.args.pathLength,
				store:          st,
				log:            logger,
			}

			// Make Shorten call
			usr, err := user.New()
			require.NoError(t, err)
			shortURL, err := svc.Shorten(ctx, usr, tt.args.url)

			// Check results
			if tt.want.err != nil {
				require.Empty(t, shortURL)
				require.ErrorIs(t, err, tt.want.err)
				return
			}
			require.NoError(t, err)

			u, err := url.Parse(shortURL)
			require.NoError(t, err)
			require.Equal(t, tt.args.pathLength, uint(len(u.Path)-1))
			require.Equal(t, u.Scheme, tt.args.servedScheme)
			require.Equal(t, u.Host, tt.args.host)

			// Check storage content
			storedURL, err := st.Get(ctx, u.Path[1:])
			require.NoError(t, err)
			assert.Equal(t, tt.args.url, storedURL)
		})
	}
}
