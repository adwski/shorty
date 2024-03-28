// Package generators implements random sequence generators.
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
	for i := range b {
		b[i] = alphabet[mrand.Intn(len(alphabet))]
	}
	return string(b)
}
