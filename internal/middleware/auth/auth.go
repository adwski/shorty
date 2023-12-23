package auth

import (
	"fmt"
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

func (mw *Middleware) HandleFunc(h http.Handler) http.Handler {
	mw.handler = h
	return mw
}
