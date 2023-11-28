package validate

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		errMsg string
	}{
		{
			name: "valid path",
			path: "/qwe123wetw4",
		},
		{
			name:   "invalid path",
			path:   "/qwe$123wetw4",
			errMsg: "invalid character in path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Path(tt.path)
			if tt.errMsg == "" {
				assert.Nil(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

func TestShortenRequest(t *testing.T) {
	type want struct {
		size   int
		errMsg string
	}
	tests := []struct {
		name string
		req  *http.Request
		want want
	}{
		{
			name: "valid Content-Length",
			req: &http.Request{
				Header: http.Header{
					"Content-Length": []string{"123"},
				},
			},
			want: want{
				size: 123,
			},
		},
		{
			name: "missing Content-Length",
			req:  &http.Request{},
			want: want{
				errMsg: "missing Content-Length",
			},
		},
		{
			name: "incorrect Content-Length",
			req: &http.Request{
				Header: http.Header{
					"Content-Length": []string{"qweasd"},
				},
			},
			want: want{
				errMsg: "incorrect Content-Length",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := ShortenRequest(tt.req)
			if tt.want.errMsg == "" {
				assert.Nil(t, err)
				assert.Equal(t, tt.want.size, size)
			} else {
				assert.Contains(t, err.Error(), tt.want.errMsg)
			}
		})
	}
}
