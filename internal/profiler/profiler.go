package profiler

import (
	"context"
	"errors"
	"net/http"
	"net/http/pprof"
	"sync"

	"go.uber.org/zap"
)

type Profiler struct {
	log *zap.Logger
	srv *http.Server
}

type Config struct {
	Logger        *zap.Logger
	ListenAddress string
}

func New(cfg *Config) *Profiler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return &Profiler{
		log: cfg.Logger.With(zap.String("component", "profiler")),
		srv: &http.Server{
			Addr:    cfg.ListenAddress,
			Handler: mux,
		},
	}
}

func (p *Profiler) Run(ctx context.Context, wg *sync.WaitGroup, errc chan error) {
	p.log.Info("starting server",
		zap.String("address", p.srv.Addr))
	errSrv := make(chan error)
	go func(errc chan error) {
		errc <- p.srv.ListenAndServe()
	}(errSrv)

	select {
	case <-ctx.Done():
		if err := p.srv.Shutdown(context.Background()); err != nil {
			p.log.Error("error during server shutdown", zap.Error(err))
		}

	case err := <-errSrv:
		if !errors.Is(err, http.ErrServerClosed) {
			p.log.Error("server error", zap.Error(err))
			errc <- err
		}
	}

	p.log.Warn("server stopped")
	wg.Done()
}
