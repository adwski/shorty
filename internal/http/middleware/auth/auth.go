// Package auth implements authentication middleware.
//
// It uses imported authenticator to handle jwt cookies.
// User object is propagated via request context.
//
// Middleware guarantees that user object will always exist in context,
// either new or parsed from cookie.
package auth

import (
	"net/http"

	"github.com/adwski/shorty/internal/user"

	authorizer "github.com/adwski/shorty/internal/auth"
	"github.com/adwski/shorty/internal/session"
	"go.uber.org/zap"
)

const (
	sessionCookieName = "shortySessID"
)

// Middleware is authentication middleware.
type Middleware struct {
	*authorizer.Auth
	handler http.Handler
	log     *zap.Logger
}

// New creates auth middleware.
func New(logger *zap.Logger, jwtSecret string) *Middleware {
	return NewFromAuthorizer(logger, authorizer.New(jwtSecret))
}

// NewFromAuthorizer creates auth middleware using existing authorizer instance.
func NewFromAuthorizer(logger *zap.Logger, a *authorizer.Auth) *Middleware {
	return &Middleware{
		Auth: a,
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

	u, token, err := mw.createOrParseUserFromRequest(r)
	if err != nil {
		logf.Error("cannot create user session", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if token != "" {
		c := &http.Cookie{
			Name:  sessionCookieName,
			Value: token,
		}
		w.Header().Set("Set-Cookie", c.String())
	}

	// Call next handler with user context
	mw.handler.ServeHTTP(w, r.WithContext(session.SetUserContext(r.Context(), u)))
}

func (mw *Middleware) createOrParseUserFromRequest(r *http.Request) (*user.User, string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return mw.CreateUserAndToken() //nolint:wrapcheck // err is checked in func above
	}
	return mw.CreateOrParseUserFromJWTString(cookie.Value) //nolint:wrapcheck // err is checked in func above
}

// HandlerFunc sets upstream middleware handler.
func (mw *Middleware) HandlerFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
}
