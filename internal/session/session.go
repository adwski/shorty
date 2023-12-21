package session

import (
	"context"

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

func GetUserFromContext(ctx context.Context) (u *user.User, ok bool) {
	u, ok = ctx.Value(ctxKeyUserID).(*user.User)
	return
}

func SetRequestID(parent context.Context, reqID string) context.Context {
	return context.WithValue(parent, ctxKeyReqID, reqID)
}

func GetRequestID(ctx context.Context) (reqID string, ok bool) {
	reqID, ok = ctx.Value(ctxKeyReqID).(string)
	return
}
