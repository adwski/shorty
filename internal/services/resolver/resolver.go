// Package resolver implements shortened URLs redirects.
// It's independent of shortener service and potentially could be used separately.
package resolver

import (
	"context"
	"errors"
	"net/http"

	"github.com/adwski/shorty/internal/session"

	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/validate"
	"go.uber.org/zap"
)

type Storage interface {
	Get(ctx context.Context, key string) (url string, err error)
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
	reqID, ok := session.GetRequestID(req.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		svc.log.Error("request id was not provided in context")
		return
	}
	logf := svc.log.With(zap.String("id", reqID))

	if err := validate.Path(req.URL.Path); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logf.Error("redirect path is not valid", zap.Error(err))
		return
	}

	redirect, err := svc.store.Get(req.Context(), req.URL.Path[1:])
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, storage.ErrDeleted):
			w.WriteHeader(http.StatusGone)
		default:
			logf.Error("cannot get redirect", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Location", redirect)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
