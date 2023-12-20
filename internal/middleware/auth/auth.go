package auth

import (
	"net/http"

	authorizer "github.com/adwski/shorty/internal/auth"
	"github.com/adwski/shorty/internal/session"
	"github.com/adwski/shorty/internal/user"
	"go.uber.org/zap"
)

type Middleware struct {
	*authorizer.Auth
	handler http.Handler
	log     *zap.Logger
}

func New(logger *zap.Logger, jwtSecret string) *Middleware {
	return &Middleware{
		Auth: authorizer.New(jwtSecret),
		log:  logger.With(zap.String("component", "auth")),
	}
}

func (mw *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-ID")
	logf := mw.log.With(zap.String("id", requestID))

	u, err := mw.GetUser(r)
	if err != nil {
		// Missing or invalid session cookie
		// Generate a new user
		logf.Debug("cannot get session from cookie", zap.Error(err))
		u, err = user.New()
		if err != nil {
			// we're helpless here
			logf.Debug("cannot create new user", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Set Cookie for new user
		sessionCookie, errS := mw.CreateJWTCookie(u)
		if errS != nil {
			logf.Debug("cannot create session cookie for user", zap.Error(errS))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Set-Cookie", sessionCookie.String())
	}

	u.SetRequestID(requestID)
	// Call next handler with user context
	mw.handler.ServeHTTP(w, r.WithContext(session.SetUserContext(r.Context(), u)))
}

func (mw *Middleware) HandleFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
}

func (mw *Middleware) Chain(h http.Handler) *Middleware {
	mw.handler = h
	return mw
}
