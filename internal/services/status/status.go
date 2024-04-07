// Package status provides simple status service indicating whether DB storage is alive or not.
package status

import (
	"context"
	"errors"

	"github.com/adwski/shorty/internal/model"
	"go.uber.org/zap"
)

// Storage is a storage type used by status service.
type Storage interface {
	Ping(context.Context) error
	Stats(context.Context) (*model.Stats, error)
}

// ErrStorageError is service error caused by underlying storage error.
var (
	ErrStorageError = errors.New("storage error")
)

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

// Ping pings the storage and returns error if ping is unsuccessful.
func (svc *Service) Ping(ctx context.Context) error {
	if err := svc.store.Ping(ctx); err != nil {
		return errors.Join(ErrStorageError, err)
	}
	return nil
}

// Stats returns storage statistics.
func (svc *Service) Stats(ctx context.Context) (*model.Stats, error) {
	stats, err := svc.store.Stats(ctx)
	if err != nil {
		return nil, errors.Join(ErrStorageError, err)
	}
	return stats, nil
}
