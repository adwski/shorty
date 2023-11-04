package validate

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {

	tests := []struct {
		name string
		path string
		err  string
	}{
		{
			name: "valid path",
			path: "/qwe123wetw4",
		},
		{
			name: "invalid path",
			path: "/qwe$123wetw4",
			err:  "invalid character in path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Path(tt.path)
			if tt.err == "" {
				assert.Nil(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.err)
			}
		})
	}
}

func TestShortenRequest(t *testing.T) {
	type want struct {
		size int
		err  string
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
				err: "missing Content-Length",
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
				err: "incorrect Content-Length",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := ShortenRequest(tt.req)
			if tt.want.err == "" {
				assert.Nil(t, err)
				assert.Equal(t, tt.want.size, size)
			} else {
				assert.Contains(t, err.Error(), tt.want.err)
			}
		})
	}
}
