package app

import (
	"github.com/adwski/shorty/internal/app/config"
	"github.com/adwski/shorty/internal/services/redirector"
	"github.com/go-chi/chi/v5"
	stdLog "log"
	"net/http"
	"time"

	"github.com/adwski/shorty/internal/services/shortener"
	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/storage/simple"
	log "github.com/sirupsen/logrus"
)

const (
	defaultTimeout = time.Second

	defaultPathLength = 8
)

type Shorty struct {
	server *http.Server
	store  storage.Storage
}

func NewShorty(cfg *config.ShortyConfig) *Shorty {

	store := simple.NewSimple(&simple.Config{PathLength: defaultPathLength})
	logW := log.StandardLogger().Writer()

	sh := &Shorty{
		store: store,
		server: &http.Server{
			Addr:              cfg.ListenAddr,
			ReadTimeout:       defaultTimeout,
			ReadHeaderTimeout: defaultTimeout,
			WriteTimeout:      defaultTimeout,
			IdleTimeout:       defaultTimeout,
			MaxHeaderBytes:    8 * http.DefaultMaxHeaderBytes,
			ErrorLog:          stdLog.New(logW, "shorty", 0),
		},
	}

	router := chi.NewRouter()
	router.Get("/{path}", redirector.New(&redirector.Config{
		Store: store,
	}).Redirect)
	router.Post("/", shortener.New(&shortener.Config{
		Store:          store,
		ServedScheme:   cfg.ServedScheme,
		RedirectScheme: cfg.RedirectScheme,
		Host:           cfg.Host,
	}).Shorten)
	router.Trace("/printDB", sh.handlePrintDB)

	sh.server.Handler = router
	return sh
}

func (sh *Shorty) Run() error {
	return sh.server.ListenAndServe()
}

func (sh *Shorty) handlePrintDB(w http.ResponseWriter, _ *http.Request) {
	log.Debug(sh.store.Dump())
	w.WriteHeader(http.StatusNoContent)
}
