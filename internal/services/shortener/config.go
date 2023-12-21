package shortener

import (
	"github.com/adwski/shorty/internal/buffer"
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
	logger := cfg.Logger.With(zap.String("component", "shortener"))

	svc := &Service{
		store:          cfg.Store,
		servedScheme:   cfg.ServedScheme,
		redirectScheme: cfg.RedirectScheme,
		host:           cfg.Host,
		pathLength:     cfg.PathLength,
		log:            logger,
	}

	svc.flusher, svc.delURLs = buffer.NewFlusher(cfg.Logger, svc.deleteURLs)
	return svc
}
