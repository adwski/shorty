package validate

import (
	"errors"
	"net/http"
	"strconv"
	"unicode"
)

func ErrInvalidChar() error {
	return errors.New("invalid character in path")
}

func ErrContentLength() error {
	return errors.New("incorrect or missing Content-Length")
}

// ShortenRequest validates http request for shorten service
func ShortenRequest(req *http.Request) (size int, err error) {
	//if req.Header.Get("Content-Type") != "text/plain" {
	//	err = errors.New("wrong Content-Type")
	//	return
	//}
	if size, err = strconv.Atoi(req.Header.Get("Content-Length")); err != nil {
		err = ErrContentLength()
		return
	}
	return
}

// Path validates http request path for redirector service
func Path(path string) error {
	for i := 1; i < len(path); i++ {
		if !unicode.IsLetter(rune(path[i])) && !unicode.IsDigit(rune(path[i])) {
			return ErrInvalidChar()
		}
	}
	return nil
}
