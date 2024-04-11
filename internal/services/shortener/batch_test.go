package shortener

import (
	"context"
	"net/url"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adwski/shorty/internal/app/mockapp"
	"github.com/adwski/shorty/internal/buffer"
	"github.com/adwski/shorty/internal/model"
	"github.com/adwski/shorty/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_ShortenBatch(t *testing.T) {
	type args struct {
		batch       []BatchURL
		serveHost   string
		serveScheme string
		pathLen     uint
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
			name: "shorten batch",
			args: args{
				serveHost:   "aaa",
				serveScheme: "http",
				pathLen:     7,
				batch: []BatchURL{
					{
						ID:  "123",
						URL: "http://qwe.qwe",
					},
					{
						ID:  "456",
						URL: "http://asd.asd",
					},
				},
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

			// Prepare mock calls
			st.EXPECT().StoreBatch(mock.Anything, mock.Anything).Once().Return(
				func(ctx context.Context, urls []model.URL) error {
					for _, u := range urls {
						st.EXPECT().Get(mock.Anything, u.Short).Return(u.Orig, nil)
					}
					return nil
				})

			// Init service
			svc := New(&Config{
				Store:        st,
				Logger:       logger,
				ServedScheme: tt.args.serveScheme,
				Host:         tt.args.serveHost,
				PathLength:   tt.args.pathLen,
			})

			usr, err := user.New()
			require.NoError(t, err)

			shortBatch, err := svc.ShortenBatch(ctx, usr, tt.args.batch)

			if tt.want.err != nil {
				assert.Nil(t, shortBatch)
				assert.ErrorIs(t, err, tt.want.err)
				return
			}
			require.NoError(t, err)

			// Check response correctness
			for i := range shortBatch {
				assert.Equal(t, tt.args.batch[i].ID, shortBatch[i].ID)
				u, errU := url.Parse(shortBatch[i].Short)
				require.NoError(t, errU)
				assert.Equal(t, int(tt.args.pathLen), len(u.Path)-1)
				assert.Equal(t, tt.args.serveHost, u.Host)
				assert.Equal(t, tt.args.serveScheme, u.Scheme)

				orig, err := st.Get(ctx, u.Path[1:])
				assert.NoError(t, err)
				assert.Equal(t, tt.args.batch[i].URL, orig)
			}
		})
	}
}

func TestService_DeleteURLs(t *testing.T) {
	type args struct {
		shorts  []string
		userID  string
		newUser bool
	}
	type want struct {
		err          error
		deletedCount int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "delete urls",
			args: args{
				shorts: []string{"qweqwe", "asdasd", "zxczxc"},
				userID: "testuser",
			},
			want: want{
				deletedCount: 3,
			},
		},
		{
			name: "delete urls new user",
			args: args{
				shorts:  []string{"qweqwe", "asdasd", "zxczxc"},
				newUser: true,
			},
			want: want{
				err:          ErrUnauthorized,
				deletedCount: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			var (
				st          = mockapp.NewStorage(t)
				wg          = sync.WaitGroup{}
				deletedURLS = make(map[string]model.URL)
				ctx, cancel = context.WithCancel(context.Background())
			)
			defer cancel()

			// spawn flusher
			flusher := buffer.NewFlusher[model.URL](&buffer.FlusherConfig{
				Logger:        logger,
				FlushInterval: 10 * time.Second,
				FlushSize:     10,
				AllocSize:     20,
			}, func(ctx context.Context, urls []model.URL) {
				for _, u := range urls {
					deletedURLS[u.Short] = u
				}
			})

			// spawn shortener
			svc := &Service{
				store:   st,
				log:     logger,
				flusher: flusher,
			}

			// run flusher
			wg.Add(1)
			go flusher.Run(ctx, &wg)

			// create user
			var u *user.User
			if tt.args.newUser {
				u, err = user.New()
				require.NoError(t, err)
			} else if tt.args.userID != "" {
				u = &user.User{ID: tt.args.userID}
			}

			// Execute
			err = svc.DeleteBatch(ctx, u, tt.args.shorts)
			if tt.want.err != nil {
				assert.ErrorIs(t, err, tt.want.err)
				return
			}
			require.NoError(t, err)

			// stop flusher
			cancel()
			wg.Wait()

			// check deleted urls
			assert.Len(t, deletedURLS, tt.want.deletedCount)
			if tt.want.deletedCount > 0 {
				// Check deleted URLs
				for _, short := range tt.args.shorts {
					_, ok := deletedURLS[short]
					assert.True(t, ok, "url must be deleted: "+short)
				}
			}
		})
	}
}

func TestService_GetURLs(t *testing.T) {
	type args struct {
		storageURLS  []*model.URL
		userID       string
		newUser      bool
		servedHost   string
		servedScheme string
	}
	type want struct {
		err  error
		urls []model.URL
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "delete urls",
			args: args{
				servedHost:   "aaa.ccc",
				servedScheme: "bbb",
				storageURLS: []*model.URL{{
					Short: "qweqwe",
					Orig:  "http://qwe.asd/zxc",
				}},
				userID: "testuser",
			},
			want: want{
				urls: []model.URL{
					{
						Short: "qweqwe",
						Orig:  "http://qwe.asd/zxc",
					},
				},
			},
		},
		{
			name: "delete urls",
			args: args{
				servedHost:   "aaa.ccc",
				servedScheme: "bbb",
				storageURLS: []*model.URL{{
					Short: "qweqwe",
					Orig:  "http://qwe.asd/zxc",
				}},
				userID: "testuser",
			},
			want: want{
				urls: []model.URL{
					{
						Short: "qweqwe",
						Orig:  "http://qwe.asd/zxc",
					},
				},
			},
		},
		{
			name: "delete urls new user",
			args: args{
				servedHost:   "aaa.ccc",
				servedScheme: "bbb",
				userID:       "testuser",
				newUser:      true,
			},
			want: want{
				err: ErrUnauthorized,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			var (
				st  = mockapp.NewStorage(t)
				ctx = context.Background()
				svc = &Service{
					store:        st,
					log:          logger,
					servedScheme: tt.args.servedScheme,
					host:         tt.args.servedHost,
				}
			)
			if len(tt.want.urls) > 0 {
				// Prepare storage mock calls
				st.EXPECT().ListUserURLs(ctx, tt.args.userID).Once().Return(tt.args.storageURLS, nil)
			}

			// Prepare user
			var usr *user.User
			if tt.args.newUser {
				usr, err = user.New()
				require.NoError(t, err)
			} else if tt.args.userID != "" {
				usr = &user.User{ID: tt.args.userID}
			}

			// Execute
			urls, err := svc.GetAll(ctx, usr)
			if tt.want.err != nil {
				assert.Nil(t, urls)
				assert.ErrorIs(t, err, tt.want.err)
				return
			}
			require.NoError(t, err)

			// Check URLs
			require.Equal(t, len(tt.want.urls), len(urls))
			sort.Slice(urls, func(i, j int) bool { return urls[i].Short > urls[j].Short })
			sort.Slice(tt.want.urls, func(i, j int) bool { return urls[i].Short > urls[j].Short })
			for i, wURL := range tt.want.urls {
				u, err := url.Parse(urls[i].Short)
				require.NoError(t, err)
				assert.Equal(t, wURL.Short, strings.TrimLeft(u.Path, "/"))
				assert.Equal(t, wURL.Orig, urls[i].Orig)
			}
		})
	}
}
