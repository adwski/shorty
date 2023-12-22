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

func SetUserContext(parent context.Context, u *user.User) context.Context {
	return context.WithValue(parent, ctxKeyUserID, u)
}

func GetUserFromContext(ctx context.Context) (*user.User, bool) {
	u, ok := ctx.Value(ctxKeyUserID).(*user.User)
	return u, ok
}

func SetRequestID(parent context.Context, reqID string) context.Context {
	return context.WithValue(parent, ctxKeyReqID, reqID)
}

func GetRequestID(ctx context.Context) (string, bool) {
	reqID, ok := ctx.Value(ctxKeyReqID).(string)
	return reqID, ok
}

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
