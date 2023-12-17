package requestid

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/gofrs/uuid/v5"
)

type Middleware struct {
	gen     uuid.Generator
	handler http.Handler
	log     *zap.Logger
}

type Config struct {
	Logger   *zap.Logger
	Generate bool
}

func New(cfg *Config) *Middleware {
	m := &Middleware{log: cfg.Logger}
	if cfg.Generate {
		m.gen = uuid.NewGen()
	}
	return m
}

func (mw *Middleware) HandlerFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
}

func (mw *Middleware) Chain(h http.Handler) *Middleware {
	mw.handler = h
	return mw
}

func (mw *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if mw.gen != nil {
		mw.setRequestID(r)
	}
	mw.handler.ServeHTTP(w, r)
}

func (mw *Middleware) setRequestID(r *http.Request) {
	u, err := mw.gen.NewV4()
	if err != nil {
		mw.log.Error("cannot generate unique request id", zap.Error(err))
		return
	}

	r.Header.Set("X-Request-ID", u.String())
}
