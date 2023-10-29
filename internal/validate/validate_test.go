package validate

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestPath(t *testing.T) {

	tests := []struct {
		name string
		path string
		err  error
	}{
		{
			name: "valid path",
			path: "/qwe123wetw4",
		},
		{
			name: "invalid path",
			path: "/qwe$123wetw4",
			err:  ErrInvalidChar(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Path(tt.path)
			if tt.err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, err.Error(), tt.err.Error())
			}
		})
	}
}

func TestShortenRequest(t *testing.T) {
	type want struct {
		size int
		err  error
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
				err: ErrContentLength(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := ShortenRequest(tt.req)
			if tt.want.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, tt.want.size, size)
			} else {
				assert.Equal(t, err.Error(), tt.want.err.Error())
			}
		})
	}
}
