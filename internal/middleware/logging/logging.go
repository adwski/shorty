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
	http.ResponseWriter
	status int
	size   int
}

func newResponseWrapper(w http.ResponseWriter) *rwWrapper {
	return &rwWrapper{
		ResponseWriter: w,
	}
}

func (rw *rwWrapper) Write(b []byte) (n int, err error) {
	n, err = rw.ResponseWriter.Write(b)
	rw.size += n
	return
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
		reqID = r.Header.Get("X-Request-ID")
	)

	mw.log.Info("request",
		zap.String("id", reqID),
		zap.String("method", r.Method),
		zap.String("uri", r.URL.Path))

	rw := newResponseWrapper(w)

	mw.handler.ServeHTTP(rw, r)

	mw.log.Info("response",
		zap.String("id", reqID),
		zap.Int("status", rw.status),
		zap.Int("size", rw.size),
		zap.Duration("duration", time.Since(start)))
}
