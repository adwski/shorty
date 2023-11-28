package resolver

import (
	"errors"
	"net/http"

	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/validate"
	"go.uber.org/zap"
)

type Storage interface {
	Get(key string) (url string, err error)
	Store(key string, url string, overwrite bool) error
}

// Service implements http handler for url redirects.
// It uses url storage as source for short urls mappings.
type Service struct {
	store Storage
	log   *zap.Logger
}

type Config struct {
	Store  Storage
	Logger *zap.Logger
}

func New(cfg *Config) *Service {
	return &Service{
		store: cfg.Store,
		log:   cfg.Logger,
	}
}

// Resolve reads URL path, retrieves corresponding URL from storage
// and returns 307 response. It performs path validation before calling storage.
func (svc *Service) Resolve(w http.ResponseWriter, req *http.Request) {
	var (
		redirect string
		err      error
	)
	if err = validate.Path(req.URL.Path); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		svc.log.Error("redirect path is not valid", zap.Error(err))
		return
	}

	if redirect, err = svc.store.Get(req.URL.Path[1:]); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			svc.log.Error("cannot get redirect", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Location", redirect)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
