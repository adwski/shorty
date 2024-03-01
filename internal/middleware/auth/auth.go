// Package auth implements authentication middleware.
//
// It uses imported authenticator to handle jwt cookies.
// User object is propagated via request context.
//
// Middleware guarantees that user object will always exist in context,
// either new or parsed from cookie.
package auth

import (
	"fmt"
	"net/http"

	authorizer "github.com/adwski/shorty/internal/auth"
	"github.com/adwski/shorty/internal/session"
	"github.com/adwski/shorty/internal/user"
	"go.uber.org/zap"
)

// Middleware is authentication middleware.
type Middleware struct {
	*authorizer.Auth
	handler http.Handler
	log     *zap.Logger
}

// New creates auth middleware.
func New(logger *zap.Logger, jwtSecret string) *Middleware {
	return &Middleware{
		Auth: authorizer.New(jwtSecret),
		log:  logger.With(zap.String("component", "auth")),
	}
}

// ServeHTTP implements auth flow for incoming request.
// If jwt cookie is present and valid, user is added to request context.
// In other cases new user is generated.
func (mw *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID, ok := session.GetRequestID(r.Context())
	if !ok {
		mw.log.Error("cannot get request id from context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logf := mw.log.With(zap.String("id", requestID))

	u, err := mw.GetUser(r)
	if err != nil {
		logf.Debug("cannot get session from cookie", zap.Error(err))
		// Missing or invalid session cookie
		// Generate a new user
		var sessionCookie *http.Cookie
		if u, sessionCookie, err = mw.createUserAndCookie(); err != nil {
			logf.Error("cannot create user session", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Set-Cookie", sessionCookie.String())
	}

	// Call next handler with user context
	mw.handler.ServeHTTP(w, r.WithContext(session.SetUserContext(r.Context(), u)))
}

func (mw *Middleware) createUserAndCookie() (*user.User, *http.Cookie, error) {
	// Create user with unique ID
	u, err := user.New()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create new user: %w", err)
	}

	// Set Cookie for new user
	sessionCookie, errS := mw.CreateJWTCookie(u)
	if errS != nil {
		return nil, nil, fmt.Errorf("cannot create session cookie for user: %w", errS)
	}
	return u, sessionCookie, nil
}

// HandlerFunc sets upstream middleware handler.
func (mw *Middleware) HandlerFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
}
