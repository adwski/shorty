// Package server implements shorty grpc server.
package server

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/adwski/shorty/internal/config"
	g "github.com/adwski/shorty/internal/grpc"
	"github.com/adwski/shorty/internal/grpc/interceptors/auth"
	"github.com/adwski/shorty/internal/grpc/interceptors/filter"
	"github.com/adwski/shorty/internal/grpc/interceptors/logging"
	"github.com/adwski/shorty/internal/grpc/interceptors/requestid"
	"github.com/adwski/shorty/internal/services/resolver"
	"github.com/adwski/shorty/internal/services/shortener"
	"github.com/adwski/shorty/internal/services/status"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

const (
	defaultServerConnectionTimeout = 15 * time.Second
)

// Server is grpc transport server for shorty app.
type Server struct {
	g.UnimplementedShortenerServer

	logger *zap.Logger

	shortenerSvc *shortener.Service
	resolverSvc  *resolver.Service
	statusSvc    *status.Service

	addr string
	opts []grpc.ServerOption

	reflection bool
}

// NewServer creates new grpc transport server.
func NewServer(
	logger *zap.Logger,
	cfg *config.Config,
	resolverSvc *resolver.Service,
	shortenerSvc *shortener.Service,
	statusSvc *status.Service,
) *Server {
	// assign options
	var opts []grpc.ServerOption
	opts = append(opts, grpc.ConnectionTimeout(defaultServerConnectionTimeout))
	if t := cfg.GetTLSConfig(); t != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(t)))
	}

	// assign interceptors
	opts = append(opts,
		// requestID
		grpc.ChainUnaryInterceptor(requestid.New(logger, cfg.TrustRequestID).Get()),
		// logging
		grpc.ChainUnaryInterceptor(logging.New(logger).Get()),
		// filter
		grpc.ChainUnaryInterceptor(filter.NewFromFilter(cfg.GetFilter(), []string{"/shorty.shortener/Stats"}).Get()),
		// auth
		grpc.ChainUnaryInterceptor(auth.NewFromAuthorizer(logger, cfg.GetAuthorizer()).Get()))

	return &Server{
		logger:       logger.With(zap.String("component", "grpc")),
		shortenerSvc: shortenerSvc,
		resolverSvc:  resolverSvc,
		statusSvc:    statusSvc,
		opts:         opts,
		addr:         cfg.GRPCListenAddr,
		reflection:   cfg.GRPCReflection,
	}
}

// Run starts grpc server and returns only when Listener stops (canceled).
// It is intended to be started asynchronously and canceled via context.
// Error channel should be used to catch listen errors.
// If error is caught that means server is no longer running.
func (srv *Server) Run(ctx context.Context, wg *sync.WaitGroup, errc chan error) {
	defer wg.Done()

	listener, err := net.Listen("tcp", srv.addr)
	if err != nil {
		srv.logger.Error("cannot create listener", zap.Error(err))
		errc <- err
		return
	}

	// create grpc server
	s := grpc.NewServer(srv.opts...)
	g.RegisterShortenerServer(s, srv)
	if srv.reflection {
		reflection.Register(s)
	}

	// start server
	srv.logger.Info("starting server", zap.String("address", listener.Addr().String()))
	errSrv := make(chan error)
	go func(errc chan<- error) {
		errc <- s.Serve(listener)
	}(errSrv)

	// wait for signals
	select {
	case <-ctx.Done():
		s.GracefulStop()
		if err = listener.Close(); err != nil {
			srv.logger.Error("error while closing listener", zap.Error(err))
		}
	case err = <-errSrv:
		srv.logger.Error("listener error", zap.Error(err))
		errc <- err
		s.GracefulStop()
	}
}
