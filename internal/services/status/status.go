// Package status provides simple status service indicating whether DB storage is alive or not.
package status

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

// Pingable is a pingable storage type used by status service.
type Pingable interface {
	Ping(context.Context) error
}

// Service is a status service.
type Service struct {
	store Pingable
	log   *zap.Logger
}

// Config is status service config.
type Config struct {
	Storage Pingable
	Logger  *zap.Logger
}

// New creates new status service.
func New(cfg *Config) *Service {
	return &Service{
		store: cfg.Storage,
		log:   cfg.Logger.With(zap.String("component", "status")),
	}
}

// Ping pings the storage and returns 200 if ping is successful or 500 otherwise.
func (svc *Service) Ping(w http.ResponseWriter, req *http.Request) {
	err := svc.store.Ping(req.Context())
	if err != nil {
		svc.log.Error("storage ping unsuccessful", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
