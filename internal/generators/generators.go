package generators

import "math/rand"

var (
	alphabet = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

// RandString generates random string with specified length
// from predefined alphabet
func RandString(length uint) string {
	b := make([]byte, length)
	for i := uint(0); i < length; i++ {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(b)
}
