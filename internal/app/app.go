package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/adwski/shorty/internal/middleware/recovering"

	"github.com/adwski/shorty/internal/app/config"
	"github.com/adwski/shorty/internal/middleware/compress"
	"github.com/adwski/shorty/internal/middleware/logging"
	"github.com/adwski/shorty/internal/middleware/requestid"
	"github.com/adwski/shorty/internal/services/resolver"
	"github.com/adwski/shorty/internal/services/shortener"
	"github.com/adwski/shorty/internal/services/status"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	defaultReadHeaderTimeout = time.Second
	defaultReadTimeout       = 5 * time.Second
	defaultWriteTimeout      = 5 * time.Second
	defaultIdleTimeout       = 10 * time.Second

	defaultPathLength = 8
)

// Shorty is URL shortener app
// It consists of shortener and redirector services
// Also it uses key-value storage to store URLs and shortened paths.
type Shorty struct {
	log    *zap.Logger
	server *http.Server
	host   string
}

// NewShorty creates Shorty instance from config.
func NewShorty(cfg *config.Shorty) *Shorty {
	shortenerSvc := shortener.New(&shortener.Config{
		Store:          cfg.Storage,
		ServedScheme:   cfg.ServedScheme,
		RedirectScheme: cfg.RedirectScheme,
		Host:           cfg.Host,
		Logger:         cfg.Logger,
		PathLength:     defaultPathLength,
	})

	resolverSvc := resolver.New(&resolver.Config{
		Store:  cfg.Storage,
		Logger: cfg.Logger,
	})

	router := chi.NewRouter()
	router.Use(
		recovering.New(&recovering.Config{Logger: cfg.Logger}).ChainFunc,
		requestid.New(&requestid.Config{
			Generate: cfg.GenerateReqID,
			Logger:   cfg.Logger,
		}).ChainFunc,
		logging.New(&logging.Config{
			Logger: cfg.Logger,
		}).ChainFunc,
		compress.New().ChainFunc,
	)

	router.Post("/", shortenerSvc.ShortenPlain)
	router.Post("/api/shorten", shortenerSvc.ShortenJSON)
	router.Post("/api/shorten/batch", shortenerSvc.ShortenBatch)
	router.Get("/{path}", resolverSvc.Resolve)

	if statusSvc, err := status.New(&status.Config{Storage: cfg.Storage}); err == nil {
		router.Get("/ping", statusSvc.PingStorage)
	} else {
		cfg.Logger.Debug("ping is not mounted", zap.Error(err))
	}

	return &Shorty{
		log:  cfg.Logger,
		host: cfg.Host,
		server: &http.Server{
			Addr:              cfg.ListenAddr,
			ReadTimeout:       defaultReadTimeout,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			WriteTimeout:      defaultWriteTimeout,
			IdleTimeout:       defaultIdleTimeout,
			ErrorLog:          log.New(newSrvLogger(cfg.Logger), "", 0),
			Handler:           router,
		},
	}
}

func (sh *Shorty) Run(ctx context.Context, wg *sync.WaitGroup, errc chan error) {
	sh.log.Info("starting server",
		zap.String("address", sh.server.Addr),
		zap.String("host", sh.host))

	errSrv := make(chan error)
	go func(errc chan error) {
		errc <- sh.server.ListenAndServe()
	}(errSrv)

	select {
	case <-ctx.Done():
		if err := sh.server.Shutdown(context.Background()); err != nil {
			sh.log.Error("error during server shutdown", zap.Error(err))
		}

	case err := <-errSrv:
		if !errors.Is(err, http.ErrServerClosed) {
			sh.log.Error("server error", zap.Error(err))
			errc <- err
		}
	}

	sh.log.Warn("server stopped")
	wg.Done()
}

type srvLogger struct {
	logger *zap.Logger
}

func newSrvLogger(logger *zap.Logger) *srvLogger {
	return &srvLogger{
		logger: logger.With(zap.String("type", "server")),
	}
}

func (sl *srvLogger) Write(b []byte) (int, error) {
	sl.logger.Error(string(b))
	return len(b), nil
}
