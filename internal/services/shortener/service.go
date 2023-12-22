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
	"sync"

	"github.com/adwski/shorty/internal/user"

	"github.com/adwski/shorty/internal/buffer"

	"github.com/adwski/shorty/internal/generators"

	"github.com/adwski/shorty/internal/storage"
	"go.uber.org/zap"
)

const (
	defaultStoreRetries = 3
)

var ErrNoUser = errors.New("middleware did not provide user context")

type Storage interface {
	Get(ctx context.Context, key string) (url string, err error)
	Store(ctx context.Context, url *storage.URL, overwrite bool) (string, error)
	StoreBatch(ctx context.Context, urls []storage.URL) error
	ListUserURLs(ctx context.Context, userid string) ([]*storage.URL, error)
	DeleteUserURLs(ctx context.Context, urls []storage.URL) (int64, error)
}

// Service implements http handler for shortened urls management.
type Service struct {
	store          Storage
	flusher        *buffer.Flusher[storage.URL]
	delURLs        chan storage.URL
	log            *zap.Logger
	servedScheme   string
	redirectScheme string
	host           string
	pathLength     uint
}

func (svc *Service) Run(ctx context.Context, wg *sync.WaitGroup) {
	svc.flusher.Run(ctx, wg)
}

func (svc *Service) getServedURL(shortPath string) string {
	return fmt.Sprintf("%s://%s/%s", svc.servedScheme, svc.host, shortPath)
}

func (svc *Service) storeURL(ctx context.Context, user *user.User, u string) (path string, err error) {
	for i := 1; i <= defaultStoreRetries; i++ {
		path = generators.RandString(svc.pathLength)

		var storedPath string
		if storedPath, err = svc.store.Store(ctx, &storage.URL{
			Short:  path,
			Orig:   u,
			UserID: user.ID,
		}, false); err != nil {
			if errors.Is(err, storage.ErrConflict) {
				path = storedPath
				return
			} else if errors.Is(err, storage.ErrAlreadyExists) {
				continue
			}
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
	if len(body) == 0 {
		err = fmt.Errorf("empty body")
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
