package logging

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
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
		log: cfg.Logger.With(zap.String("component", "http")),
	}
}

type rwWrapper struct {
	http.ResponseWriter
	writeErr    error
	status      int
	size        int
	wroteHeader bool
}

func newResponseWrapper(w http.ResponseWriter) *rwWrapper {
	return &rwWrapper{
		ResponseWriter: w,
	}
}

func (rw *rwWrapper) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.flushHeader()
	}
	if rw.writeErr != nil {
		return 0, fmt.Errorf("write error occurred: %w", rw.writeErr)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	rw.writeErr = err
	return n, err //nolint:wrapcheck // pass unmodified write errors to handlers
}

func (rw *rwWrapper) flushHeader() {
	if rw.status == 0 {
		// WriteHeader was never called
		rw.status = http.StatusOK
	} else if rw.status < 100 || rw.status > 999 {
		// Incorrect code, we're preventing panic on net/http level
		rw.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		rw.writeErr = errors.New("incorrect response code")
	}
	// Delayed header write
	rw.ResponseWriter.WriteHeader(rw.status)
	rw.wroteHeader = true
}

func (rw *rwWrapper) WriteHeader(statusCode int) {
	rw.status = statusCode
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
	var (
		start = time.Now()
		reqID = r.Header.Get("X-Request-ID")
	)

	mw.log.Info("request",
		zap.String("id", reqID),
		zap.String("method", r.Method),
		zap.String("uri", r.URL.Path))

	rw := newResponseWrapper(w)

	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		mw.log.Error("panic in handler chain",
			zap.Any("panic", rec),
			zap.String("stack", string(debug.Stack())),
			zap.String("id", reqID),
			zap.Duration("duration", time.Since(start)))

		if !rw.wroteHeader {
			// All this trickery is to minimize the possibility of panic w/o 500 response
			rw.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		}
	}()

	mw.handler.ServeHTTP(rw, r)
	if !rw.wroteHeader {
		rw.flushHeader()
	}

	l := mw.log.With(zap.String("id", reqID),
		zap.Int("status", rw.status),
		zap.Int("size", rw.size),
		zap.Duration("duration", time.Since(start)))

	if rw.writeErr != nil {
		l = l.With(zap.String("writeError", rw.writeErr.Error()))
	}
	l.Info("response")
}
