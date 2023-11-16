package shortener

import (
	"github.com/adwski/shorty/internal/storage"
	"go.uber.org/zap"
)

type Config struct {
	Store          storage.Storage
	ServedScheme   string
	RedirectScheme string
	Host           string
	Logger         *zap.Logger
	PathLength     uint
}

func New(cfg *Config) *Service {
	return &Service{
		store:          cfg.Store,
		servedScheme:   cfg.ServedScheme,
		redirectScheme: cfg.RedirectScheme,
		host:           cfg.Host,
		pathLength:     cfg.PathLength,
		log:            cfg.Logger,
	}
}
