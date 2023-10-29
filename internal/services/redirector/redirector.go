package redirector

import (
	"errors"
	"net/http"

	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/storage/common"
	"github.com/adwski/shorty/internal/validate"
	log "github.com/sirupsen/logrus"
)

// Service is redirector service
type Service struct {
	store storage.Storage
}

type ServiceConfig struct {
	Store storage.Storage
}

func NewService(cfg *ServiceConfig) *Service {
	return &Service{
		store: cfg.Store,
	}
}

func (svc *Service) Redirect(w http.ResponseWriter, req *http.Request) {
	var (
		redirect string
		err      error
	)
	if err = validate.Path(req.URL.Path); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithError(err).Error("redirect path is not valid")
		return
	}

	redirect, err = svc.store.Get(req.URL.Path[1:])
	if err != nil {
		if errors.Is(err, common.ErrNotFound()) {
			w.WriteHeader(http.StatusBadRequest) // or NotFound may be?
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Location", redirect)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
