package requestid

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
)

type Middleware struct {
	gen     uuid.Generator
	handler http.Handler
}

type Config struct {
	Generate bool
}

func New(cfg *Config) *Middleware {
	m := &Middleware{}
	if cfg.Generate {
		m.gen = uuid.NewGen()
	}
	return m
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
		// error will happen only if ReadFull() fails
		// In what cases it might be so?
		return
	}

	r.Header.Set("X-Request-ID", u.String())
}
