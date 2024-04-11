package shortener

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/adwski/shorty/internal/model"

	"github.com/adwski/shorty/internal/user"

	"github.com/adwski/shorty/internal/buffer"

	"github.com/adwski/shorty/internal/generators"

	"go.uber.org/zap"
)

const (
	defaultStoreRetries = 3
)

// Service errors.
var (
	ErrInvalidURL           = errors.New("invalid url")
	ErrUnsupportedURLScheme = errors.New("unsupported scheme")
	ErrStorageError         = errors.New("storage error")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrDelete               = errors.New("cannot queue url for deletion")
	ErrEmptyBatch           = errors.New("empty batch")
)

// Storage is URL storage used by shortener.
type Storage interface {
	Get(ctx context.Context, key string) (url string, err error)
	Store(ctx context.Context, url *model.URL, overwrite bool) (string, error)
	StoreBatch(ctx context.Context, urls []model.URL) error
	ListUserURLs(ctx context.Context, userid string) ([]*model.URL, error)
	DeleteUserURLs(ctx context.Context, urls []model.URL) (int64, error)
}

// Service implements http handler for shortened urls management.
type Service struct {
	store          Storage
	flusher        *buffer.Flusher[model.URL]
	log            *zap.Logger
	servedScheme   string
	redirectScheme string
	host           string
	pathLength     uint
}

// GetFlusher returns flusher instance.
func (svc *Service) GetFlusher() *buffer.Flusher[model.URL] {
	return svc.flusher
}

// Shorten generates short URL for incoming original URL and returns short url back.
func (svc *Service) Shorten(ctx context.Context, user *user.User, origURL string) (string, error) {
	u, err := url.Parse(origURL)
	if err != nil {
		return "", errors.Join(ErrInvalidURL, err)
	}
	if svc.redirectScheme != "" && u.Scheme != svc.redirectScheme {
		return "", ErrUnsupportedURLScheme
	}

	shortPath, err := svc.storeURL(ctx, user, u.String())
	if err != nil {
		if !errors.Is(model.ErrConflict, err) {
			return "", errors.Join(ErrStorageError, err)
		}
	}
	return svc.getServedURL(shortPath), err // nil or conflict
}

func (svc *Service) getServedURL(shortPath string) string {
	return fmt.Sprintf("%s://%s/%s", svc.servedScheme, svc.host, shortPath)
}

func (svc *Service) storeURL(ctx context.Context, user *user.User, u string) (path string, err error) {
	for i := 1; i <= defaultStoreRetries; i++ {
		path = generators.RandString(svc.pathLength)

		var storedPath string
		if storedPath, err = svc.store.Store(ctx, &model.URL{
			Short:  path,
			Orig:   u,
			UserID: user.ID,
		}, false); err != nil {
			if errors.Is(err, model.ErrConflict) {
				path = storedPath
				return
			} else if errors.Is(err, model.ErrAlreadyExists) {
				continue
			}
		}
		return
	}
	err = fmt.Errorf("cannot store url: %w", err)
	return
}
