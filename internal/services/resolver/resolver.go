package resolver

import (
	"net/http"

	"github.com/adwski/shorty/internal/errors"
	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/validate"
	"github.com/sirupsen/logrus"
)

// Service is resolver service
type Service struct {
	store storage.Storage
	log   *logrus.Logger
}

type Config struct {
	Store  storage.Storage
	Logger *logrus.Logger
}

func New(cfg *Config) *Service {
	return &Service{
		store: cfg.Store,
		log:   cfg.Logger,
	}
}

// Resolve reads URL path, retrieves corresponding URL from storage
// and returns 307 response. It performs path validation before calling storage
func (svc *Service) Resolve(w http.ResponseWriter, req *http.Request) {
	var (
		redirect string
		err      error
	)
	if err = validate.Path(req.URL.Path); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		svc.log.WithError(err).Error("redirect path is not valid")
		return
	}

	if redirect, err = svc.store.Get(req.URL.Path[1:]); err != nil {
		if errors.Equal(err, errors.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			svc.log.WithError(err).Error("cannot get redirect")
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Location", redirect)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
