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

// RandStringOldV3 generates random string with specified length
// from predefined alphabet
func RandStringOldV3(length uint) string {
	var (
		b  = make([]byte, length)
		ln = len(alphabet)
	)
	for i := uint(0); i < length; i++ {
		b[i] = alphabet[rand.Intn(ln)]
	}
	return string(b)
}

// RandStringOld generates random string with specified length
// from alphanumeric characters.
func RandStringOld(length uint) string {
	var (
		ch int
		i  uint
		b  = make([]byte, length)
	)
	for i = 0; i < length; i++ {
		ch = rand.Intn(62)
		if ch < 10 {
			// [0-9]
			b[i] = byte(ch + 48)
		} else if ch < 36 {
			// [A-Z]
			b[i] = byte(ch + 55)
		} else {
			// [a-z]
			b[i] = byte(ch + 61)
		}
	}
	return string(b)
}

// RandStringOldV2 generates random string with specified length
// from alphanumeric characters.
func RandStringOldV2(length uint) string {
	b := make([]byte, length)
	for i := uint(0); i < length; i++ {
		ch := rand.Intn(62)
		if ch < 10 {
			// [0-9]
			b[i] = byte(ch + 48)
		} else if ch < 36 {
			// [A-Z]
			b[i] = byte(ch + 55)
		} else {
			// [a-z]
			b[i] = byte(ch + 61)
		}
	}
	return string(b)
}
