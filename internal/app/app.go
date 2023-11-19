package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/adwski/shorty/internal/middleware/compress"
	"github.com/adwski/shorty/internal/middleware/logging"
	"github.com/adwski/shorty/internal/middleware/requestid"
	"github.com/adwski/shorty/internal/storage/file"
	"github.com/adwski/shorty/internal/storage/simple"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/adwski/shorty/internal/app/config"
	"github.com/adwski/shorty/internal/services/resolver"
	"github.com/adwski/shorty/internal/services/shortener"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	defaultReadHeaderTimeout = time.Second
	defaultReadTimeout       = 5 * time.Second
	defaultWriteTimeout      = 5 * time.Second
	defaultIdleTimeout       = 10 * time.Second

	defaultShutdownTimeout = 10 * time.Second

	defaultPathLength = 8
)

type Storage interface {
	Get(key string) (url string, err error)
	Store(key string, url string, overwrite bool) error
}

// Shorty is URL shortener app
// It consists of shortener and redirector services
// Also it uses key-value storage to store URLs and shortened paths
type Shorty struct {
	log    *zap.Logger
	server *http.Server
	host   string
	ctx    context.Context
}

// NewShorty creates Shorty instance from config
func NewShorty(ctx context.Context, cfg *config.ShortyConfig) (*Shorty, error) {

	//--------------------------------------------------
	// Create URL storage
	//--------------------------------------------------
	var (
		storage Storage
		err     error
	)
	switch cfg.Storage {
	case config.StorageKindSimple:
		storage = simple.New()
	case config.StorageKindFile:
		if storage, err = file.New(&file.Config{
			FilePath: cfg.FileStoragePath,
			Ctx:      ctx,
			Logger:   cfg.Logger,
		}); err != nil {
			return nil, fmt.Errorf("cannot initialize file storage: %w", err)
		}
	}

	shortenerSvc := shortener.New(&shortener.Config{
		Store:          storage,
		ServedScheme:   cfg.ServedScheme,
		RedirectScheme: cfg.RedirectScheme,
		Host:           cfg.Host,
		Logger:         cfg.Logger,
		PathLength:     defaultPathLength,
	})

	resolverSvc := resolver.New(&resolver.Config{
		Store:  storage,
		Logger: cfg.Logger,
	})

	router := chi.NewRouter()
	router.Post("/", shortenerSvc.ShortenPlain)
	router.Post("/api/shorten", shortenerSvc.ShortenJSON)
	router.Get("/{path}", resolverSvc.Resolve)

	chain := requestid.New(&requestid.Config{Generate: cfg.GenerateReqID}).Chain(
		logging.New(&logging.Config{Logger: cfg.Logger}).Chain(
			compress.New().Chain(router)))

	return &Shorty{
		ctx:  ctx,
		log:  cfg.Logger,
		host: cfg.Host,
		server: &http.Server{
			Addr:              cfg.ListenAddr,
			ReadTimeout:       defaultReadTimeout,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			WriteTimeout:      defaultWriteTimeout,
			IdleTimeout:       defaultIdleTimeout,
			ErrorLog:          log.New(newSrvLogger(cfg.Logger), "", 0),
			Handler:           chain,
		},
	}, nil
}

func (sh *Shorty) Run(wg *sync.WaitGroup, errc chan error) {
	sh.log.Info("starting app",
		zap.String("address", sh.server.Addr),
		zap.String("host", sh.host))

	errSrv := make(chan error)
	go func(errc chan error) {
		errc <- sh.server.ListenAndServe()
	}(errSrv)

	var (
		shutdown bool
		done     = make(chan struct{})
	)

Loop:
	for {
		select {
		case <-sh.ctx.Done():
			shutdown = true
			sh.log.Warn("got canceled, shutting down")
			go func() {
				err := sh.server.Shutdown(context.Background())
				if err != nil {
					sh.log.Error("error during shutdown", zap.Error(err))
				}
				done <- struct{}{}
			}()
			break Loop

		case err := <-errSrv:
			if !errors.Is(err, http.ErrServerClosed) {
				sh.log.Error("server error", zap.Error(err))
				errc <- err
			}
		}
	}

	if shutdown {
		select {
		case <-done:
			sh.log.Warn("shutdown complete")
		case <-time.After(defaultShutdownTimeout):
			sh.log.Warn("shutdown timeout")
		}
	}

	sh.log.Warn("app stopped")
	wg.Done()
}

type srvLogger struct {
	logger *zap.Logger
}

func newSrvLogger(logger *zap.Logger) *srvLogger {
	return &srvLogger{logger: logger}
}

func (sl *srvLogger) Write(b []byte) (int, error) {
	sl.logger.Error(string(b), zap.String("type", "server"))
	return len(b), nil
}
