// Package app is a complete URL shortener backend application.
// It uses storage component that should be initialized beforehand.
package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/adwski/shorty/internal/app/config"
	"github.com/adwski/shorty/internal/middleware/auth"
	"github.com/adwski/shorty/internal/middleware/compress"
	"github.com/adwski/shorty/internal/middleware/logging"
	"github.com/adwski/shorty/internal/middleware/requestid"
	"github.com/adwski/shorty/internal/services/resolver"
	"github.com/adwski/shorty/internal/services/shortener"
	"github.com/adwski/shorty/internal/services/status"
	"github.com/adwski/shorty/internal/storage"
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

// Storage defines storage backend methods that is used by Shorty.
type Storage interface {
	Get(ctx context.Context, key string) (url string, err error)
	Store(ctx context.Context, url *storage.URL, overwrite bool) (string, error)
	StoreBatch(ctx context.Context, urls []storage.URL) error
	ListUserURLs(ctx context.Context, userid string) ([]*storage.URL, error)
	DeleteUserURLs(ctx context.Context, urls []storage.URL) (int64, error)
	Ping(ctx context.Context) error
	Close()
}

// Shorty is URL shortener app
// It consists of shortener and redirector services
// Also it uses key-value storage to store URLs and shortened paths.
type Shorty struct {
	log       *zap.Logger
	server    *http.Server
	shortener *shortener.Service
	host      string
}

// NewShorty creates Shorty instance from config.
func NewShorty(logger *zap.Logger, storage Storage, cfg *config.Shorty) *Shorty {
	shortenerSvc := shortener.New(&shortener.Config{
		Store:          storage,
		ServedScheme:   cfg.ServedScheme,
		RedirectScheme: cfg.RedirectScheme,
		Host:           cfg.Host,
		Logger:         logger,
		PathLength:     defaultPathLength,
	})
	resolverSvc := resolver.New(&resolver.Config{
		Store:  storage,
		Logger: logger,
	})
	statusSvc := status.New(&status.Config{
		Storage: storage,
		Logger:  logger,
	})

	r := getRouterWithMiddleware(logger, cfg.TrustRequestID)
	r.With(auth.New(logger, cfg.JWTSecret).HandleFunc).Route("/", func(r chi.Router) {
		r.Get("/api/user/urls", shortenerSvc.GetURLs)
		r.Delete("/api/user/urls", shortenerSvc.DeleteURLs)
		r.Post("/api/shorten", shortenerSvc.ShortenJSON)
		r.Post("/api/shorten/batch", shortenerSvc.ShortenBatch)
		r.Post("/", shortenerSvc.ShortenPlain)
	})
	r.Get("/{path}", resolverSvc.Resolve)
	r.Get("/ping", statusSvc.Ping)

	return &Shorty{
		log:       logger.With(zap.String("component", "api")),
		host:      cfg.Host,
		shortener: shortenerSvc,
		server: &http.Server{
			Addr:              cfg.ListenAddr,
			ReadTimeout:       defaultReadTimeout,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			WriteTimeout:      defaultWriteTimeout,
			IdleTimeout:       defaultIdleTimeout,
			ErrorLog:          log.New(newSrvLogger(logger), "", 0),
			Handler:           r,
		},
	}
}

// Run starts internal web server and returned only wen ListenAndServe returns.
// It is intended to be started asynchronously and canceled via context.
// Error channel should be used to catch listen errors.
// If error is caught that means web server is no longer running.
func (sh *Shorty) Run(ctx context.Context, wg *sync.WaitGroup, errc chan error) {
	wg.Add(1)
	go sh.shortener.Run(ctx, wg)

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

func getRouterWithMiddleware(logger *zap.Logger, trustRequestID bool) chi.Router {
	router := chi.NewRouter()
	router.Use(
		requestid.New(&requestid.Config{
			Trust:  trustRequestID,
			Logger: logger,
		}).HandlerFunc,
		logging.New(&logging.Config{
			Logger: logger,
		}).HandlerFunc,
		compress.New().HandlerFunc,
	)
	return router
}

type srvLogger struct {
	logger *zap.Logger
}

func newSrvLogger(logger *zap.Logger) *srvLogger {
	return &srvLogger{
		logger: logger.With(zap.String("component", "server")),
	}
}

func (sl *srvLogger) Write(b []byte) (int, error) {
	sl.logger.Error(string(b))
	return len(b), nil
}
