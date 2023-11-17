package app

import (
	"github.com/adwski/shorty/internal/middleware/compress"
	"github.com/adwski/shorty/internal/middleware/logging"
	"github.com/adwski/shorty/internal/middleware/requestid"
	"net/http"
	"time"

	"github.com/adwski/shorty/internal/app/config"
	"github.com/adwski/shorty/internal/services/resolver"
	"github.com/adwski/shorty/internal/services/shortener"
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

// Shorty is URL shortener app
// It consists of shortener and redirector services
// Also it uses key-value storage to store URLs and shortened paths
type Shorty struct {
	log    *zap.Logger
	server *http.Server
	host   string
}

// NewShorty creates Shorty instance from config
func NewShorty(cfg *config.ShortyConfig) *Shorty {

	var (
		store  = storage.NewStorageSimple()
		router = chi.NewRouter()
	)

	shortenerSvc := shortener.New(&shortener.Config{
		Store:          store,
		ServedScheme:   cfg.ServedScheme,
		RedirectScheme: cfg.RedirectScheme,
		Host:           cfg.Host,
		Logger:         cfg.Logger,
		PathLength:     defaultPathLength,
	})

	resolverSvc := resolver.New(&resolver.Config{
		Store:  store,
		Logger: cfg.Logger,
	})

	router.Post("/", shortenerSvc.ShortenPlain)
	router.Post("/api/shorten", shortenerSvc.ShortenJSON)

	router.Get("/{path}", resolverSvc.Resolve)

	return &Shorty{
		log:  cfg.Logger,
		host: cfg.Host,
		server: &http.Server{
			Addr:              cfg.ListenAddr,
			ReadTimeout:       defaultReadTimeout,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			WriteTimeout:      defaultWriteTimeout,
			IdleTimeout:       defaultIdleTimeout,
			ErrorLog:          zap.NewStdLog(cfg.Logger),

			Handler: requestid.New(&requestid.Config{Generate: cfg.GenerateReqID}).Chain(
				logging.New(&logging.Config{Logger: cfg.Logger}).Chain(
					compress.New().Chain(router))),
		},
	}
}

func (sh *Shorty) Run() error {
	sh.log.Info("starting app",
		zap.String("address", sh.server.Addr),
		zap.String("host", sh.host))

	return sh.server.ListenAndServe()
}
