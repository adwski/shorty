package redirector

import (
	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/storage/common"
	"github.com/adwski/shorty/internal/validate"
	log "github.com/sirupsen/logrus"

	"errors"
	"net/http"
)

// Service is redirector service
type Service struct {
	store  storage.Storage
	scheme string
}

type ServiceConfig struct {
	Store  storage.Storage
	Scheme string
}

func NewService(cfg *ServiceConfig) *Service {
	return &Service{
		store:  cfg.Store,
		scheme: cfg.Scheme,
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
		if errors.Is(err, common.ErrErrorNotFound()) {
			w.WriteHeader(http.StatusBadRequest) // or NotFound may be?
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Location", svc.scheme+"://"+redirect)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
