package session

import (
	"context"

	"github.com/adwski/shorty/internal/user"
)

type ctxKey int

const (
	ctxKeyUID ctxKey = iota
	ctxReqID
)

func SetUserContext(parent context.Context, u *user.User) context.Context {
	return context.WithValue(parent, ctxKeyUID, u)
}

func GetUserFromContext(ctx context.Context) (u *user.User, ok bool) {
	u, ok = ctx.Value(ctxKeyUID).(*user.User)
	return
}

func SetRequestID(parent context.Context, reqID string) context.Context {
	return context.WithValue(parent, ctxReqID, reqID)
}

func GetRequestID(ctx context.Context) (reqID string, ok bool) {
	reqID, ok = ctx.Value(ctxReqID).(string)
	return
}
