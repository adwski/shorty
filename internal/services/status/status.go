package status

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type Pingable interface {
	Ping(context.Context) error
}

type Service struct {
	store Pingable
	log   *zap.Logger
}

type Config struct {
	Storage Pingable
	Logger  *zap.Logger
}

func New(cfg *Config) *Service {
	return &Service{
		store: cfg.Storage,
		log:   cfg.Logger.With(zap.String("component", "status")),
	}
}

func (svc *Service) Ping(w http.ResponseWriter, req *http.Request) {
	err := svc.store.Ping(req.Context())
	if err != nil {
		svc.log.Error("storage ping unsuccessful", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
