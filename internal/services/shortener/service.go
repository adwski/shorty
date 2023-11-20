package shortener

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/adwski/shorty/internal/errors"
	"github.com/adwski/shorty/internal/generators"
)

const (
	defaultMaxTries = 10
)

type Storage interface {
	Get(key string) (url string, err error)
	Store(key string, url string, overwrite bool) error
}

// Service implements http handler for shortened urls management.
type Service struct {
	store          Storage
	log            *zap.Logger
	servedScheme   string
	redirectScheme string
	host           string
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

func getRedirectURLFromBody(req *http.Request) (u *url.URL, err error) {
	var body []byte
	if body, err = readBody(req); err != nil {
		err = fmt.Errorf("cannot get url from request body: %w", err)
		return
	}
	if u, err = url.Parse(string(body)); err != nil {
		err = fmt.Errorf("cannot parse url from request body: %w", err)
	}
	return
}

func readBody(req *http.Request) (body []byte, err error) {
	var (
		r io.ReadCloser
	)
	defer func() { _ = req.Body.Close() }()

	if r, err = getContentReader(req); err != nil {
		return
	}
	defer func() { _ = r.Close() }()

	body, err = io.ReadAll(r)
	return
}

func getContentReader(req *http.Request) (r io.ReadCloser, err error) {
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
