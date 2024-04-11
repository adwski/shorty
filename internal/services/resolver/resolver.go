// Package resolver implements shortened URLs redirects.
// It's independent of shortener service and potentially could be used separately.
package resolver

import (
	"context"
	"errors"
	"fmt"
	"unicode"

	"go.uber.org/zap"
)

// Service errors.
var (
	ErrInvalidPath  = errors.New("invalid path")
	ErrStorageError = errors.New("storage error")
)

// Storage is URL storage used by resolver.
type Storage interface {
	Get(ctx context.Context, key string) (url string, err error)
}

// Service implements http handler for url redirects.
// It uses url storage as source for short urls mappings.
type Service struct {
	store Storage
	log   *zap.Logger
}

// Config is resolver service config.
type Config struct {
	Store  Storage
	Logger *zap.Logger
}

// New creates new resolver service.
func New(cfg *Config) *Service {
	return &Service{
		store: cfg.Store,
		log:   cfg.Logger,
	}
}

// Resolve lookups original URL using incoming shortened path.
func (svc *Service) Resolve(ctx context.Context, path string) (string, error) {
	if err := validatePath(path); err != nil {
		return "", errors.Join(ErrInvalidPath, err)
	}
	origURL, err := svc.store.Get(ctx, path[1:])
	if err != nil {
		return "", errors.Join(ErrStorageError, err)
	}
	return origURL, nil
}

func validatePath(path string) error {
	if path[0] != '/' {
		return fmt.Errorf("path is not starts with /")
	}
	for i := 1; i < len(path); i++ {
		if !unicode.IsLetter(rune(path[i])) && !unicode.IsDigit(rune(path[i])) {
			return fmt.Errorf("invalid character in path: 0x%x", path[i])
		}
	}
	return nil
}
