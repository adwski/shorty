package compress

import (
	"compress/gzip"
	"compress/zlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stub struct {
	body        []byte
	contentType string
}

func (s *stub) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", s.contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(s.body)
}

func TestMiddleware(t *testing.T) {
	type args struct {
		acceptEncoding  string
		respContentType string
	}
	type want struct {
		contentEncoding string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "accept gzip response non-text",
			args: args{
				acceptEncoding:  "gzip",
				respContentType: "what/ever",
			},
			want: want{
				contentEncoding: "",
			},
		},
		{
			name: "accept gzip response json",
			args: args{
				acceptEncoding:  "gzip",
				respContentType: "application/json",
			},
			want: want{
				contentEncoding: "gzip",
			},
		},
		{
			name: "accept gzip response plain",
			args: args{
				acceptEncoding:  "gzip",
				respContentType: "text/plain",
			},
			want: want{
				contentEncoding: "gzip",
			},
		},
		{
			name: "not accept gzip response json",
			args: args{
				acceptEncoding:  "",
				respContentType: "application/json",
			},
			want: want{
				contentEncoding: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := New()

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.args.acceptEncoding != "" {
				r.Header.Set("Accept-Encoding", tt.args.acceptEncoding)
			}
			s := &stub{
				body:        []byte("zxczxczxc"),
				contentType: tt.args.respContentType,
			}
			mw.Chain(s)

			w := httptest.NewRecorder()

			mw.ServeHTTP(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			respEnc := resp.Header.Get("Content-Encoding")
			require.Equal(t, tt.want.contentEncoding, respEnc)

			var (
				body io.ReadCloser
				err  error
			)
			switch respEnc {
			case "gzip":
				body, err = gzip.NewReader(resp.Body)
			case "deflate":
				body, err = zlib.NewReader(resp.Body)
			default:
				body = resp.Body
			}
			require.Nil(t, err)
			defer body.Close()

			bodyContent, errb := io.ReadAll(body)
			require.Nil(t, errb)

			assert.Equal(t, s.body, bodyContent)
		})
	}
}
