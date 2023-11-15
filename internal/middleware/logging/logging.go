package logging

import (
	"net/http"
	"time"

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

type rwWrapper struct {
	status int
	size   int
	http.ResponseWriter
}

func newResponseWrapper(w http.ResponseWriter) *rwWrapper {
	return &rwWrapper{
		ResponseWriter: w,
	}
}

func (rw *rwWrapper) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

func (rw *rwWrapper) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *rwWrapper) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (mw *Middleware) Chain(h http.Handler) *Middleware {
	mw.handler = h
	return mw
}

func (mw *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var (
		start = time.Now()
		reqId = r.Header.Get("X-Request-ID")
	)

	mw.log.Info("request",
		zap.String("id", reqId),
		zap.String("method", r.Method),
		zap.String("uri", r.URL.Path))

	rw := newResponseWrapper(w)

	mw.handler.ServeHTTP(rw, r)

	mw.log.Info("response",
		zap.String("id", reqId),
		zap.Int("status", rw.status),
		zap.Int("size", rw.size),
		zap.Duration("duration", time.Since(start)))
}
