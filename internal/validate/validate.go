package validate

import (
	"errors"
	"net/http"
	"strconv"
	"unicode"
)

var (
	ErrInvalidChar            = errors.New("invalid character in path")
	ErrContentLengthMissing   = errors.New("missing Content-Length")
	ErrContentLengthIncorrect = errors.New("incorrect Content-Length")
)

// ShortenRequest validates http request for shorten service
func ShortenRequest(req *http.Request) (size int, err error) {
	cl := req.Header.Get("Content-Length")
	if cl == "" {
		err = ErrContentLengthMissing
		return
	}
	if size, err = strconv.Atoi(cl); err != nil {
		err = errors.Join(ErrContentLengthIncorrect, err)
	}
	return
}

// Path validates http request path for redirector service
func Path(path string) error {
	for i := 1; i < len(path); i++ {
		if !unicode.IsLetter(rune(path[i])) && !unicode.IsDigit(rune(path[i])) {
			return ErrInvalidChar
		}
	}
	return nil
}
