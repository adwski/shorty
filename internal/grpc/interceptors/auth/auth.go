// Package auth implements authentication interceptor.
//
// It uses imported authenticator to handle jwt metadata parameters.
// User object is propagated via request context.
//
// Interceptor guarantees that user object will always exist in context,
// either new or parsed from cookie.
//
//nolint:wrapcheck // return grpc errors
package auth

import (
	"context"

	authorizer "github.com/adwski/shorty/internal/auth"
	"github.com/adwski/shorty/internal/session"
	"github.com/adwski/shorty/internal/user"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	gstatus "google.golang.org/grpc/status"
)

const (
	sessionKey = "shorty-sess-id"
)

// Interceptor is an auth interceptor.
type Interceptor struct {
	*authorizer.Auth
	logger *zap.Logger
}

// NewFromAuthorizer creates auth interceptor using exiting authorizer.
func NewFromAuthorizer(logger *zap.Logger, a *authorizer.Auth) *Interceptor {
	return &Interceptor{
		Auth:   a,
		logger: logger.With(zap.String("component", "auth-interceptor")),
	}
}

// Get returns UnaryServerInterceptor func.
func (i *Interceptor) Get() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			i.logger.Error("cannot get request metadata")
			return nil, gstatus.Error(codes.Internal, "cannot get metadata")
		}
		u, token, err := i.createOrParseUserFromMetadata(md)
		if err != nil {
			i.logger.Error("cannot create user session")
			return nil, gstatus.Error(codes.Internal, "cannot create user session")
		}
		if token != "" {
			if err = grpc.SendHeader(ctx, metadata.Pairs(
				sessionKey, token,
			)); err != nil {
				i.logger.Error("cannot send header")
				return nil, gstatus.Error(codes.Internal, "cannot send header")
			}
		}
		return handler(session.SetUserContext(ctx, u), req)
	}
}

func (i *Interceptor) createOrParseUserFromMetadata(md metadata.MD) (*user.User, string, error) {
	val := md.Get(sessionKey)
	if len(val) == 0 {
		return i.CreateUserAndToken()
	}
	return i.CreateOrParseUserFromJWTString(val[0])
}
