package requestid

import (
	"net/http"

	"github.com/adwski/shorty/internal/session"

	"go.uber.org/zap"

	"github.com/gofrs/uuid/v5"
)

type Middleware struct {
	gen     uuid.Generator
	handler http.Handler
	log     *zap.Logger
	trust   bool
}

type Config struct {
	Logger *zap.Logger
	Trust  bool
}

func New(cfg *Config) *Middleware {
	return &Middleware{
		log:   cfg.Logger.With(zap.String("component", "request-id")),
		gen:   uuid.NewGen(),
		trust: cfg.Trust,
	}
}

func (mw *Middleware) HandlerFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
}

func (mw *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if mw.trust {
		if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
			mw.handler.ServeHTTP(w, r.WithContext(session.SetRequestID(r.Context(), reqID)))
			return
		}
		mw.log.Debug("incoming request without id but trust is enabled")
	}

	u, err := mw.gen.NewV4()
	if err != nil {
		mw.log.Error("cannot generate unique request id", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	mw.handler.ServeHTTP(w, r.WithContext(session.SetRequestID(r.Context(), u.String())))
}
