package generators

import (
	crand "crypto/rand"
	mrand "math/rand"
)

var (
	alphabet = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

// RandStringFallback generates random string with specified length
// from predefined alphabet. It uses math/rand's weak rng.
func RandStringFallback(length uint) string {
	b := make([]byte, length)
	for i := uint(0); i < length; i++ {
		b[i] = alphabet[mrand.Intn(len(alphabet))]
	}
	return string(b)
}

// RandString generates random string with specified length
// from predefined alphabet.
func RandString(length uint) string {
	b := make([]byte, length)

	if _, err := crand.Read(b); err != nil {
		return RandStringFallback(length)
	}
	for i := 0; i < int(length); i++ {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}
