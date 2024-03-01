package shortener

import (
	"time"

	"github.com/adwski/shorty/internal/buffer"
	"go.uber.org/zap"
)

const (
	flusherFillSize      = 100
	flusherAllocSize     = 200
	flusherFlushInterval = 3 * time.Second
)

// Config is shortener service configuration.
type Config struct {
	Store          Storage
	Logger         *zap.Logger
	ServedScheme   string
	RedirectScheme string
	Host           string
	PathLength     uint
}

// New create new shortener service.
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

	svc.flusher = buffer.NewFlusher(&buffer.FlusherConfig{
		Logger:        cfg.Logger,
		FlushInterval: flusherFlushInterval,
		FlushSize:     flusherFillSize,
		AllocSize:     flusherAllocSize,
	}, svc.deleteURLs)
	return svc
}
