// Package requestid implements request ID middleware.
//
// Request IDs are generated using UUIDv4. Middleware has an option
// to trust incoming X-Request-ID headers.
package requestid

import (
	"context"
	"net/http"

	"github.com/adwski/shorty/internal/session"

	"go.uber.org/zap"

	"github.com/gofrs/uuid/v5"
)

// Middleware is requestID middleware.
type Middleware struct {
	gen     uuid.Generator
	handler http.Handler
	log     *zap.Logger
	trust   bool
}

// Config is requestID middleware configuration.
type Config struct {
	Logger *zap.Logger
	Trust  bool
}

// New creates requestID middleware.
func New(cfg *Config) *Middleware {
	return &Middleware{
		log:   cfg.Logger.With(zap.String("component", "request-id")),
		gen:   uuid.NewGen(),
		trust: cfg.Trust,
	}
}

// HandlerFunc sets upstream middleware handler.
func (mw *Middleware) HandlerFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
}

// ServeHTTP checks requestID header and/or generates new requestID which is passed
// to upstream handlers using request context.
func (mw *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var newCtx context.Context
	if mw.trust {
		if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
			newCtx = session.SetRequestID(r.Context(), reqID)
		}
		mw.log.Debug("incoming request without id but trust is enabled")
	}

	if newCtx == nil {
		u, err := mw.gen.NewV4()
		if err != nil {
			mw.log.Error("cannot generate unique request id", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		newCtx = session.SetRequestID(r.Context(), u.String())
	}
	mw.handler.ServeHTTP(w, r.WithContext(newCtx))
}
