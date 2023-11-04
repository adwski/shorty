package validate

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"unicode"
)

// ShortenRequest validates http request for shorten service
func ShortenRequest(req *http.Request) (size int, err error) {
	cl := req.Header.Get("Content-Length")
	if cl == "" {
		err = errors.New("missing Content-Length")
		return
	}
	if size, err = strconv.Atoi(cl); err != nil {
		err = errors.Join(errors.New("incorrect Content-Length"), err)
	}
	return
}

// Path validates http request path for redirector service
func Path(path string) error {
	for i := 1; i < len(path); i++ {
		if !unicode.IsLetter(rune(path[i])) && !unicode.IsDigit(rune(path[i])) {
			return fmt.Errorf("invalid character in path: 0x%x", path[i])
		}
	}
	return nil
}
