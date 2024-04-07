package resolver

import (
	"context"
	"testing"

	"github.com/adwski/shorty/internal/app/mockapp"
	"github.com/adwski/shorty/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_Redirect(t *testing.T) {
	type args struct {
		pathLength   uint
		path         string
		invalid      bool
		addToStorage map[string]string
	}
	type want struct {
		orig string
		err  error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "redirect existing",
			args: args{
				pathLength: 10,
				path:       "/qweasdzxcr",
				addToStorage: map[string]string{
					"qweasdzxcr": "https://aaa.bbb",
				},
			},
			want: want{
				orig: "https://aaa.bbb",
			},
		},
		{
			name: "redirect existing, different path length",
			args: args{
				pathLength: 20,
				path:       "/qweasdzxcr",
				addToStorage: map[string]string{
					"qweasdzxcr": "https://aaa.bbb",
				},
			},
			want: want{
				orig: "https://aaa.bbb",
			},
		},
		{
			name: "redirect not existing",
			args: args{
				pathLength: 10,
				path:       "/qweasd1xcr",
				addToStorage: map[string]string{
					"qweasdzxcr": "https://aaa.bbb",
				},
			},
			want: want{
				err: model.ErrNotFound,
			},
		},
		{
			name: "invalid request path",
			args: args{
				pathLength: 10,
				path:       "/qweasd&*zxcrаб",
				invalid:    true,
				addToStorage: map[string]string{
					"qweasdzxcr123": "https://aaa.bbb",
				},
			},
			want: want{
				err: ErrInvalidPath,
			},
		},
		{
			name: "invalid request path 2",
			args: args{
				pathLength: 10,
				path:       "qweasdzxcr123",
				invalid:    true,
				addToStorage: map[string]string{
					"qweasdzxcr123": "https://aaa.bbb",
				},
			},
			want: want{
				err: ErrInvalidPath,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			st := mockapp.NewStorage(t)

			if v, ok := tt.args.addToStorage[tt.args.path[1:]]; !ok {
				if !tt.args.invalid {
					st.EXPECT().Get(mock.Anything, tt.args.path[1:]).Return("", model.ErrNotFound)
				}
			} else {
				st.EXPECT().Get(mock.Anything, tt.args.path[1:]).Return(v, nil)
			}

			svc := &Service{
				store: st,
				log:   logger,
			}

			orig, err := svc.Resolve(context.Background(), tt.args.path)
			if tt.want.err != nil {
				assert.Empty(t, orig)
				assert.ErrorIs(t, err, tt.want.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.orig, orig)
			}
		})
	}
}
