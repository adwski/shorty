package recovering

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

type Middleware struct {
	log     *zap.Logger
	handler http.Handler
}

type Config struct {
	Logger *zap.Logger
}

func New(cfg *Config) *Middleware {
	return &Middleware{
		log: cfg.Logger,
	}
}

func (mw *Middleware) ChainFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
}

func (mw *Middleware) Chain(h http.Handler) *Middleware {
	mw.handler = h
	return mw
}

func (mw *Middleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		mw.log.Error("panic in handler chain",
			zap.Any("panic", r),
			zap.String("stack", string(debug.Stack())))

		// Doesn't work here for some reason
		w.WriteHeader(http.StatusInternalServerError)
	}()

	mw.handler.ServeHTTP(w, req)
}
