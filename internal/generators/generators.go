package generators

import (
	mrand "math/rand"
)

var (
	alphabet = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

// RandString generates random string with specified length
// from predefined alphabet.
func RandString(length uint) string {
	b := make([]byte, length)
	for i := uint(0); i < length; i++ {
		b[i] = alphabet[mrand.Intn(len(alphabet))]
	}
	return string(b)
}
