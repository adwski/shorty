// Package server contains http server implementation for shorty app.
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/adwski/shorty/internal/config"
	"github.com/adwski/shorty/internal/http/middleware/auth"
	"github.com/adwski/shorty/internal/http/middleware/compress"
	"github.com/adwski/shorty/internal/http/middleware/filter"
	"github.com/adwski/shorty/internal/http/middleware/logging"
	"github.com/adwski/shorty/internal/http/middleware/requestid"
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
)

// Server is http server instance.
// It includes router and pointers to business logic handlers.
type Server struct {
	logger       *zap.Logger
	shortenerSvc *shortener.Service
	resolverSvc  *resolver.Service
	statusSvc    *status.Service
	tls          *tls.Config
	hSrv         *http.Server
}

// NewServer creates Server instance.
func NewServer(
	logger *zap.Logger,
	cfg *config.Config,
	resolverSvc *resolver.Service,
	shortenerSvc *shortener.Service,
	statusSvc *status.Service,
) *Server {
	srv := &Server{
		logger:       logger.With(zap.String("component", "httpserver")),
		resolverSvc:  resolverSvc,
		shortenerSvc: shortenerSvc,
		statusSvc:    statusSvc,
		tls:          cfg.GetTLSConfig(),
	}
	var (
		router     = getRouterWithMiddleware(logger, cfg.TrustRequestID)
		authorizer = auth.NewFromAuthorizer(logger, cfg.GetAuthorizer())
		filterMW   = filter.NewFromFilter(cfg.GetFilter())
	)
	srv.registerHandlers(router, authorizer, filterMW)
	srv.hSrv = &http.Server{
		TLSConfig:         cfg.GetTLSConfig(),
		Addr:              cfg.ListenAddr,
		ReadTimeout:       defaultReadTimeout,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
		ErrorLog:          log.New(newSrvLogger(logger), "", 0),
		Handler:           router,
	}
	return srv
}

// Handler returns root of handler chain of underlying http.Server.
func (srv *Server) Handler() http.Handler {
	return srv.hSrv.Handler
}

// Run starts http server and returned only when ListenAndServe returns.
// It is intended to be started asynchronously and canceled via context.
// Error channel should be used to catch listen errors.
// If error is caught that means server is no longer running.
func (srv *Server) Run(ctx context.Context, wg *sync.WaitGroup, errc chan error) {
	srv.logger.Info("starting server", zap.String("address", srv.hSrv.Addr))
	errSrv := make(chan error)
	go func(errc chan error) {
		if srv.hSrv.TLSConfig != nil {
			// cert and key are provided via tls.Config
			errc <- srv.hSrv.ListenAndServeTLS("", "")
		} else {
			errc <- srv.hSrv.ListenAndServe()
		}
	}(errSrv)
	select {
	case <-ctx.Done():
		if err := srv.hSrv.Shutdown(context.Background()); err != nil {
			srv.logger.Error("error during server shutdown", zap.Error(err))
		}
	case err := <-errSrv:
		if !errors.Is(err, http.ErrServerClosed) {
			srv.logger.Error("server error", zap.Error(err))
			errc <- err
		}
	}
	srv.logger.Warn("server stopped")
	wg.Done()
}

func (srv *Server) registerHandlers(r chi.Router, authMW *auth.Middleware, filterMW *filter.Middleware) {
	r.With(authMW.HandlerFunc).Route("/", func(r chi.Router) {
		r.Get("/api/user/urls", srv.GetAll)
		r.Delete("/api/user/urls", srv.DeleteBatch)
		r.Post("/api/shorten", srv.Shorten)
		r.Post("/api/shorten/batch", srv.ShortenBatch)
		r.Post("/", srv.ShortenPlain)
	})
	r.Get("/{path}", srv.Resolve)
	r.Get("/ping", srv.Ping)
	r.With(filterMW.HandlerFunc).Get("/api/internal/stats", srv.Stats)
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
		logger: logger.With(zap.String("component", "httpserver")),
	}
}

// Write writes byte slice as one Error log message.
func (sl *srvLogger) Write(b []byte) (int, error) {
	sl.logger.Error(string(b))
	return len(b), nil
}
