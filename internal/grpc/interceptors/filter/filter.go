// Package filter contains filtering interceptor.
// When included in chain, middleware matches incoming request
// address against configured trusted subnets and passes request
// to upstream handler only in case of successful match.
// If request address doesn't match trusted subnets,
// PermissionDenied code is sent back.
//
// Interceptor has additional list of methods for which
// it should perform request filtering. Requests to all other methods
// will be passed unchecked.
//
//nolint:wrapcheck // return grpc errors
package filter

import (
	"context"

	"github.com/adwski/shorty/internal/filter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	gstatus "google.golang.org/grpc/status"
)

// Interceptor is filtering interceptor.
type Interceptor struct {
	*filter.Filter
	methods []string
	block   bool
}

// NewFromFilter creates interceptor using exiting filter instance.
func NewFromFilter(f *filter.Filter, methods []string) *Interceptor {
	return &Interceptor{
		Filter:  f,
		methods: methods,
		block:   len(f.Subnets()) == 0,
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
		if !i.methodMatch(info) {
			return handler(ctx, req)
		}

		if i.block {
			return nil, gstatus.Error(codes.PermissionDenied, "denied")
		}

		var (
			remoteAddr, xff, xRealIP string
		)
		if p, ok := peer.FromContext(ctx); ok {
			remoteAddr = p.Addr.String()
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if xffS := md.Get("x-forwarded-for"); len(xffS) > 0 {
				xff = xffS[0]
			}
			if xRealIPS := md.Get("x-real-ip"); len(xRealIPS) > 0 {
				xRealIP = xRealIPS[0]
			}
		}

		if i.CheckRequestParams(remoteAddr, xRealIP, xff) {
			return handler(ctx, req)
		}
		return nil, gstatus.Error(codes.PermissionDenied, "denied")
	}
}

func (i *Interceptor) methodMatch(info *grpc.UnaryServerInfo) bool {
	for _, m := range i.methods {
		if m == info.FullMethod {
			return true
		}
	}
	return false
}
