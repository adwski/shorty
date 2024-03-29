// Package status provides simple status service indicating whether DB storage is alive or not.
package status

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/adwski/shorty/internal/model"

	"go.uber.org/zap"
)

// Storage is a storage type used by status service.
type Storage interface {
	Ping(context.Context) error
	Stats(context.Context) (*model.StatsResponse, error)
}

// Service is a status service.
type Service struct {
	store Storage
	log   *zap.Logger
}

// Config is status service config.
type Config struct {
	Storage Storage
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

func (svc *Service) Stats(w http.ResponseWriter, req *http.Request) {
	stats, err := svc.store.Stats(req.Context())
	if err != nil {
		svc.log.Error("cannot get storage stats", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(stats)
	if err != nil {
		svc.log.Error("cannot marshal stats response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(b); err != nil {
		svc.log.Error("cannot write stats body")
	}
}
