// Package requestid implements request ID interceptor.
//
// Request IDs are generated using UUIDv4. Interceptor has an option
// to trust incoming x-request-id metadata param.
//
//nolint:wrapcheck // return grpc errors
package requestid

import (
	"context"

	"github.com/adwski/shorty/internal/session"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	gstatus "google.golang.org/grpc/status"
)

// Interceptor is request-id interceptor.
type Interceptor struct {
	logger *zap.Logger
	gen    uuid.Generator
	trust  bool
}

// New creates request-id interceptor.
func New(logger *zap.Logger, trustRequestID bool) *Interceptor {
	return &Interceptor{
		gen:    uuid.NewGen(),
		logger: logger.With(zap.String("component", "request-id-interceptor")),
		trust:  trustRequestID,
	}
}

// Get returns UnaryServerInterceptor func that can be used for chaining.
func (i *Interceptor) Get() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var newCtx context.Context
		if i.trust {
			reqID, ok := getRequestIDFromContext(ctx)
			if ok {
				newCtx = session.SetRequestID(ctx, reqID)
			}
			i.logger.Debug("incoming request without id but trust is enabled")
		}

		if newCtx == nil {
			uuidV4, err := i.gen.NewV4()
			if err != nil {
				i.logger.Error("cannot generate request id", zap.Error(err))
				return nil, gstatus.Error(codes.Internal, "cannot generate request id")
			}
			newCtx = session.SetRequestID(ctx, uuidV4.String())
		}

		return handler(newCtx, req)
	}
}

func getRequestIDFromContext(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	reqID := md.Get("x-request-id")
	if len(reqID) == 0 {
		return "", false
	}
	return reqID[0], true
}
