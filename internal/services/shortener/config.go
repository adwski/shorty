package shortener

import (
	"go.uber.org/zap"
)

type Config struct {
	Store          Storage
	Logger         *zap.Logger
	ServedScheme   string
	RedirectScheme string
	Host           string
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
