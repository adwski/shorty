package shortener

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/adwski/shorty/internal/errors"
	"github.com/adwski/shorty/internal/generators"
	"go.uber.org/zap"
)

const (
	defaultMaxTries = 10
)

type Storage interface {
	Get(key string) (url string, err error)
	Store(key string, url string, overwrite bool) error
}

// Service is a shortener service
type Service struct {
	store          Storage
	servedScheme   string
	redirectScheme string
	host           string
	log            *zap.Logger
	pathLength     uint
}

func (svc *Service) getServedURL(shortPath string) string {
	return fmt.Sprintf("%s://%s/%s", svc.servedScheme, svc.host, shortPath)
}

func (svc *Service) storeURL(u string) (path string, err error) {
	for try := 0; try <= defaultMaxTries; try++ {
		if try == defaultMaxTries {
			err = errors.ErrGiveUP
			return
		}
		path = generators.RandString(svc.pathLength)
		if err = svc.store.Store(path, u, false); err != nil {
			if errors.Equal(err, errors.ErrAlreadyExists) {
				continue
			}
			return
		}
		break
	}
	return
}

func getRedirectURLFromBody(req *http.Request) (*url.URL, error) {
	body, err := readBody(req)
	if err != nil {
		return nil, err
	}
	return url.Parse(string(body))
}

func readBody(req *http.Request) (body []byte, err error) {
	var (
		r io.ReadCloser
	)
	defer func() { _ = req.Body.Close() }()

	if r, err = readBodyContent(req); err != nil {
		return
	}
	defer func() { _ = r.Close() }()

	body, err = io.ReadAll(r)
	return
}

func readBodyContent(req *http.Request) (r io.ReadCloser, err error) {
	switch req.Header.Get("Content-Encoding") {
	case "gzip":
		r, err = gzip.NewReader(req.Body)
	case "deflate":
		r, err = zlib.NewReader(req.Body)
	case "":
		r = req.Body
	default:
		err = errors.ErrUnknownEncoding
	}
	return
}
