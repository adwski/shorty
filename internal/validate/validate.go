package validate

import (
	"errors"
	"net/http"
	"strconv"
	"unicode"
)

func ShortenRequest(req *http.Request) (size int, err error) {
	//if req.Header.Get("Content-Type") != "text/plain" {
	//	err = errors.New("wrong Content-Type")
	//	return
	//}
	if size, err = strconv.Atoi(req.Header.Get("Content-Length")); err != nil {
		err = errors.New("incorrect Content-Length")
		return
	}
	return
}

func Path(path string) (err error) {
	for i := 1; i < len(path); i++ {
		if !unicode.IsLetter(rune(path[i])) && !unicode.IsDigit(rune(path[i])) {
			err = errors.New("invalid character in path")
			return
		}
	}
	return
}
