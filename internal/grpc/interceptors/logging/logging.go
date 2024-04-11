// Package logging implements logging interceptor.
//
// It writes two log messages: for request and response which can be correlated using request ID.
package logging

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/adwski/shorty/internal/session"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Interceptor is logging interceptor.
type Interceptor struct {
	logger *zap.Logger
}

// New creates logging interceptor using exiting zap logger.
func New(logger *zap.Logger) *Interceptor {
	return &Interceptor{
		logger: logger.With(zap.String("component", "grpc")),
	}
}

// Get returns UnaryServerInterceptor that can be used for chaining.
func (i *Interceptor) Get() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var (
			start    = time.Now()
			reqID, _ = session.GetRequestID(ctx)
		)

		i.logger.Info("request",
			zap.String("id", reqID),
			zap.String("method", info.FullMethod))

		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			i.logger.Error("panic in handler chain",
				zap.Any("panic", rec),
				zap.String("stack", string(debug.Stack())),
				zap.String("id", reqID),
				zap.Duration("duration", time.Since(start)))
		}()

		resp, err := handler(ctx, req)

		i.logger.With(zap.Error(err),
			zap.String("id", reqID),
			zap.Duration("duration", time.Since(start))).Info("response")

		return resp, err
	}
}
