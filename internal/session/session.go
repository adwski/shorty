// Package session implements request context operation for context propagated data.
package session

import (
	"context"
	"errors"

	"github.com/adwski/shorty/internal/user"
)

type ctxKey int

const (
	ctxKeyUserID ctxKey = iota
	ctxKeyReqID
)

// SetUserContext assigns user object to request context.
func SetUserContext(parent context.Context, u *user.User) context.Context {
	return context.WithValue(parent, ctxKeyUserID, u)
}

// GetUserFromContext retrieves user object from request context.
func GetUserFromContext(ctx context.Context) (*user.User, bool) {
	u, ok := ctx.Value(ctxKeyUserID).(*user.User)
	return u, ok
}

// SetRequestID assigns request ID to request context.
func SetRequestID(parent context.Context, reqID string) context.Context {
	return context.WithValue(parent, ctxKeyReqID, reqID)
}

// GetRequestID retrieves request ID from request context.
func GetRequestID(ctx context.Context) (string, bool) {
	reqID, ok := ctx.Value(ctxKeyReqID).(string)
	return reqID, ok
}

// GetUserAndReqID retrieves both request ID and user object from request context.
func GetUserAndReqID(ctx context.Context) (*user.User, string, error) {
	userID, okU := ctx.Value(ctxKeyUserID).(*user.User)
	if !okU {
		return nil, "", errors.New("user was not provided in context")
	}
	reqID, okR := ctx.Value(ctxKeyReqID).(string)
	if !okR {
		return nil, "", errors.New("request id was not provided in context")
	}
	return userID, reqID, nil
}
