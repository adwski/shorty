package shortener

import (
	"compress/gzip"
	"compress/zlib"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/adwski/shorty/internal/generators"

	"github.com/adwski/shorty/internal/storage"
	"go.uber.org/zap"
)

const (
	defaultStoreRetries = 3
)

type Storage interface {
	Get(ctx context.Context, key string) (url string, err error)
	Store(ctx context.Context, key string, url string, overwrite bool) (string, error)
	StoreBatch(ctx context.Context, keys []string, urls []string) error
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

func (svc *Service) storeURL(ctx context.Context, u string) (path string, err error) {
	for i := 1; i <= defaultStoreRetries; i++ {
		path = generators.RandString(svc.pathLength)

		svc.log.Debug("storing url",
			zap.String("key", path),
			zap.Int("try", i),
			zap.String("url", u))

		var storedPath string
		if storedPath, err = svc.store.Store(ctx, path, u, false); err != nil {
			if errors.Is(err, storage.ErrConflict) {
				path = storedPath
				return
			}
			continue
		}
		return
	}
	err = fmt.Errorf("cannot store url: %w", err)
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
		err = errors.New("unknown content encoding")
	}
	return
}
