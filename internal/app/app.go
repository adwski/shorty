package app

import (
	stdLog "log"
	"net/http"
	"time"

	"github.com/adwski/shorty/internal/services/redirector"
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
	shortenerSvc  *shortener.Service
	redirectorSvc *redirector.Service
	server        *http.Server
	store         storage.Storage
}

type ShortyConfig struct {
	ListenAddr     string
	Host           string
	RedirectScheme string
	ServedScheme   string
}

func NewShorty(cfg *ShortyConfig) *Shorty {

	store := simple.NewSimple(&simple.Config{PathLength: defaultPathLength})
	logW := log.StandardLogger().Writer()

	sh := &Shorty{
		store: store,
		shortenerSvc: shortener.NewService(&shortener.ServiceConfig{
			Store:          store,
			ServedScheme:   cfg.ServedScheme,
			RedirectScheme: cfg.RedirectScheme,
			Host:           cfg.Host,
		}),
		redirectorSvc: redirector.NewService(&redirector.ServiceConfig{
			Store: store,
		}),
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

	sh.server.Handler = sh
	return sh
}

func (sh *Shorty) Run() error {
	return sh.server.ListenAndServe()
}

func (sh *Shorty) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		sh.redirectorSvc.Redirect(w, req)

	case http.MethodPost:
		sh.shortenerSvc.Shorten(w, req)

	case http.MethodTrace:
		sh.handleDebug(w, req)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (sh *Shorty) handleDebug(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/printDB" {
		log.Debug(sh.store.Dump())
	}
	w.WriteHeader(http.StatusNoContent)
}
