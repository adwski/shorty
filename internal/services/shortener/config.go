package shortener

import (
	"github.com/adwski/shorty/internal/buffer"
	"github.com/adwski/shorty/internal/storage"
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

	chForURLsDeletion := make(chan *storage.URL)

	svc := &Service{
		store: cfg.Store,

		delURLs:        chForURLsDeletion,
		servedScheme:   cfg.ServedScheme,
		redirectScheme: cfg.RedirectScheme,
		host:           cfg.Host,
		pathLength:     cfg.PathLength,
		log:            logger,
	}

	svc.flusher = buffer.NewFlusher(cfg.Logger, chForURLsDeletion, svc.deleteURLs)

	return svc
}
