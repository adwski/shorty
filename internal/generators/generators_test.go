package generators

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandString(t *testing.T) {
	tests := []struct {
		name   string
		length uint
		want   int
	}{
		{
			name:   "empty",
			length: 0,
			want:   0,
		},
		{
			name:   "short",
			length: 10,
			want:   10,
		},
		{
			name:   "long",
			length: 1000,
			want:   1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := RandString(tt.length)
			assert.Equal(t, tt.want, len(gen))

			if tt.length > 0 {
				assert.True(t, regexp.MustCompile("^[0-9a-zA-Z]+$").MatchString(gen))
			}
		})
	}
}

func BenchmarkRandString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = RandString(100)
	}
}
