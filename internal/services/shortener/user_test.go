package shortener

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adwski/shorty/internal/app/mockapp"
	"github.com/adwski/shorty/internal/buffer"
	"github.com/adwski/shorty/internal/session"
	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_DeleteURLs(t *testing.T) {
	type args struct {
		shorts  []string
		userID  string
		newUser bool
	}
	type want struct {
		status       int
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
				status:       http.StatusAccepted,
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
				status:       http.StatusUnauthorized,
				deletedCount: 0,
			},
		},
		{
			name: "delete urls no user",
			args: args{
				shorts: []string{"qweqwe", "asdasd", "zxczxc"},
			},
			want: want{
				status:       http.StatusInternalServerError,
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
				ctx, cancel = context.WithCancel(context.Background())
				wg          = sync.WaitGroup{}
				deletedURLS = make(map[string]storage.URL)
				flusher     = buffer.NewFlusher[storage.URL](&buffer.FlusherConfig{
					Logger:        logger,
					FlushInterval: 10 * time.Second,
					FlushSize:     10,
					AllocSize:     20,
				}, func(ctx context.Context, urls []storage.URL) {
					for _, u := range urls {
						deletedURLS[u.Short] = u
					}
				})
				svc = &Service{
					store:   st,
					log:     logger,
					flusher: flusher,
				}
			)
			wg.Add(1)
			go flusher.Run(ctx, &wg)

			// Prepare request
			body, errJ := json.Marshal(&tt.args.shorts)
			require.NoError(t, errJ)
			r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
			r = r.WithContext(session.SetRequestID(r.Context(), "testreqest"))
			if tt.args.newUser {
				u, errU := user.New()
				require.NoError(t, errU)
				r = r.WithContext(session.SetUserContext(r.Context(), u))
			} else if tt.args.userID != "" {
				r = r.WithContext(session.SetUserContext(r.Context(), &user.User{ID: tt.args.userID}))
			}
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Content-Length", strconv.Itoa(len(body)))
			w := httptest.NewRecorder()

			// Execute
			svc.DeleteURLs(w, r)
			res := w.Result()

			// Check status code
			assert.Equal(t, tt.want.status, res.StatusCode)

			// Check body
			resBody, errB := io.ReadAll(res.Body)
			require.NoError(t, errB)
			require.NoError(t, res.Body.Close())
			require.Len(t, resBody, 0)

			cancel()
			wg.Wait()

			require.Len(t, deletedURLS, tt.want.deletedCount)

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
		storageURLS  []*storage.URL
		userID       string
		newUser      bool
		servedHost   string
		servedScheme string
	}
	type want struct {
		status int
		urls   []storage.URL
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
				storageURLS: []*storage.URL{{
					Short: "qweqwe",
					Orig:  "http://qwe.asd/zxc",
				}},
				userID: "testuser",
			},
			want: want{
				status: http.StatusOK,
				urls: []storage.URL{
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
				storageURLS: []*storage.URL{{
					Short: "qweqwe",
					Orig:  "http://qwe.asd/zxc",
				}},
				userID: "testuser",
			},
			want: want{
				status: http.StatusOK,
				urls: []storage.URL{
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
				status: http.StatusUnauthorized,
			},
		},
		{
			name: "delete urls no user",
			args: args{
				servedHost:   "aaa.ccc",
				servedScheme: "bbb",
			},
			want: want{
				status: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			var (
				st  = mockapp.NewStorage(t)
				svc = &Service{
					store:        st,
					log:          logger,
					servedScheme: tt.args.servedScheme,
					host:         tt.args.servedHost,
				}
			)
			if len(tt.want.urls) > 0 {
				// Prepare storage mock calls
				st.EXPECT().ListUserURLs(mock.Anything, tt.args.userID).Once().Return(tt.args.storageURLS, nil)
			}

			// Prepare request
			r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			r = r.WithContext(session.SetRequestID(r.Context(), "testreqest"))
			if tt.args.newUser {
				u, errU := user.New()
				require.NoError(t, errU)
				r = r.WithContext(session.SetUserContext(r.Context(), u))
			} else if tt.args.userID != "" {
				r = r.WithContext(session.SetUserContext(r.Context(), &user.User{ID: tt.args.userID}))
			}
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			svc.GetURLs(w, r)
			res := w.Result()

			// Check status code
			assert.Equal(t, tt.want.status, res.StatusCode)

			// Check body
			resBody, errB := io.ReadAll(res.Body)
			require.NoError(t, errB)
			require.NoError(t, res.Body.Close())

			if len(tt.want.urls) == 0 {
				require.Equal(t, 0, len(resBody))
			} else if len(tt.want.urls) > 0 {
				require.NotEqual(t, 0, len(resBody))

				// Get URLs from json
				var urls []storage.URL
				err = json.Unmarshal(resBody, &urls)
				require.NoError(t, err)

				// Check URLs
				require.Equal(t, len(tt.want.urls), len(urls))
				sort.Slice(urls, func(i, j int) bool { return urls[i].Short > urls[j].Short })
				sort.Slice(tt.want.urls, func(i, j int) bool { return urls[i].Short > urls[j].Short })
				for i, wURL := range tt.want.urls {
					u, errU := url.Parse(urls[i].Short)
					require.NoError(t, errU)
					assert.Equal(t, wURL.Short, strings.TrimLeft(u.Path, "/"))
					assert.Equal(t, wURL.Orig, urls[i].Orig)
				}
			}
		})
	}
}
