// Package server implements shorty grpc server.
package server

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/adwski/shorty/internal/config"
	g "github.com/adwski/shorty/internal/grpc"
	"github.com/adwski/shorty/internal/services/resolver"
	"github.com/adwski/shorty/internal/services/shortener"
	"github.com/adwski/shorty/internal/services/status"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	tls  *tls.Config
	addr string
}

// NewServer creates new grpc transport server.
func NewServer(
	logger *zap.Logger,
	cfg *config.Config,
	resolverSvc *resolver.Service,
	shortenerSvc *shortener.Service,
	statusSvc *status.Service,
) *Server {
	return &Server{
		logger:       logger.With(zap.String("component", "grpc")),
		shortenerSvc: shortenerSvc,
		resolverSvc:  resolverSvc,
		statusSvc:    statusSvc,
		tls:          cfg.GetTLSConfig(),
		addr:         cfg.GRPCListenAddr,
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

	// assign options
	var opts []grpc.ServerOption
	opts = append(opts, grpc.ConnectionTimeout(defaultServerConnectionTimeout))
	if srv.tls != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(srv.tls)))
	}

	// create grpc server
	s := grpc.NewServer(opts...)
	g.RegisterShortenerServer(s, srv)

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
