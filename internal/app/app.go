// Package app is a complete URL shortener backend application.
// It uses storage component that should be initialized beforehand.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/adwski/shorty/internal/config"
	grpcserver "github.com/adwski/shorty/internal/grpc/server"
	httpserver "github.com/adwski/shorty/internal/http/server"
	"github.com/adwski/shorty/internal/model"
	"github.com/adwski/shorty/internal/profiler"
	"github.com/adwski/shorty/internal/services/resolver"
	"github.com/adwski/shorty/internal/services/shortener"
	"github.com/adwski/shorty/internal/services/status"
	"github.com/adwski/shorty/internal/storage/database"
	"github.com/adwski/shorty/internal/storage/file"
	"github.com/adwski/shorty/internal/storage/memory"
	"go.uber.org/zap"
)

const (
	defaultPathLength = 8
)

// Storage defines storage backend methods that is used by Shorty.
type Storage interface {
	Get(ctx context.Context, key string) (url string, err error)
	Store(ctx context.Context, url *model.URL, overwrite bool) (string, error)
	StoreBatch(ctx context.Context, urls []model.URL) error
	ListUserURLs(ctx context.Context, userid string) ([]*model.URL, error)
	DeleteUserURLs(ctx context.Context, urls []model.URL) (int64, error)
	Ping(ctx context.Context) error
	Stats(ctx context.Context) (*model.Stats, error)
	Close()
}

// Shorty is URL shortener app
// It consists of shortener and redirector services
// Also it uses key-value storage to store URLs and shortened paths.
type Shorty struct {
	logger       *zap.Logger
	http         *httpserver.Server
	grpc         *grpcserver.Server
	shortenerSvc *shortener.Service
}

// NewShorty creates Shorty instance from config.
func NewShorty(logger *zap.Logger, storage Storage, cfg *config.Config) (*Shorty, error) {
	if storage == nil {
		return nil, fmt.Errorf("nil storage")
	}
	shortenerSvc := shortener.New(&shortener.Config{
		Store:          storage,
		ServedScheme:   cfg.ServedScheme,
		RedirectScheme: cfg.RedirectScheme,
		Host:           cfg.ServedHost,
		Logger:         logger,
		PathLength:     defaultPathLength,
	})
	resolverSvc := resolver.New(&resolver.Config{
		Store:  storage,
		Logger: logger,
	})
	statusSvc := status.New(&status.Config{
		Storage: storage,
		Logger:  logger,
	})

	sh := &Shorty{
		logger:       logger,
		shortenerSvc: shortenerSvc,
	}
	if cfg.ListenAddr != "" {
		sh.http = httpserver.NewServer(logger, cfg, resolverSvc, shortenerSvc, statusSvc)
	}
	if cfg.GRPCListenAddr != "" {
		sh.grpc = grpcserver.NewServer(logger, cfg, resolverSvc, shortenerSvc, statusSvc)
	}
	return sh, nil
}

// Run creates and starts shorty app using provided logger and config.
// It runs until interrupted or some error occurs.
// Non-zero error code is returned in latter case.
func Run(logger *zap.Logger, cfg *config.Config) int {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	// Creating storage
	store, err := createStorage(ctx, logger, cfg.Storage)
	if err != nil {
		logger.Error("storage init error", zap.Error(err))
		return 1
	}

	var (
		wg   = &sync.WaitGroup{}
		errc = make(chan error)
	)

	// Creating app
	shorty, err := NewShorty(logger, store, cfg)
	if err != nil {
		logger.Error("cannot create app", zap.Error(err))
		return 1
	}

	if cfg.PprofServerAddr != "" {
		// creating and starting pprof server
		prof := profiler.New(&profiler.Config{
			Logger:        logger,
			ListenAddress: cfg.PprofServerAddr,
		})
		wg.Add(1)
		go prof.Run(ctx, wg, errc)
	}

	// starting flusher
	wg.Add(1)
	go shorty.shortenerSvc.GetFlusher().Run(ctx, wg)

	// starting http server
	if shorty.http != nil {
		wg.Add(1)
		go shorty.http.Run(ctx, wg, errc)
	}

	// starting grpc server
	if shorty.grpc != nil {
		wg.Add(1)
		go shorty.grpc.Run(ctx, wg, errc)
	}

	// waiting for signals
	code := 0
	select {
	case <-ctx.Done():
		logger.Warn("shutting down")
	case <-errc:
		logger.Error("caught server error, finishing", zap.Error(err))
		code = 1
		cancel()
	}
	wg.Wait()
	store.Close()
	return code
}

func createStorage(ctx context.Context, logger *zap.Logger, cfg *config.Storage) (store Storage, err error) {
	switch {
	case cfg.DatabaseDSN != "":
		if store, err = database.New(ctx, &database.Config{
			Logger:  logger,
			DSN:     cfg.DatabaseDSN,
			Migrate: true,
			Trace:   cfg.TraceDB,
		}); err != nil {
			err = fmt.Errorf("cannot initialize database storage: %w", err)
			break
		}
		logger.Debug("using DB storage")

	case cfg.FileStoragePath != "":
		if store, err = file.New(ctx, &file.Config{
			FilePath: cfg.FileStoragePath,
			Logger:   logger,
		}); err != nil {
			err = fmt.Errorf("cannot initialize file storage: %w", err)
			break
		}
		logger.Debug("using file storage")

	default:
		store = memory.New()
		logger.Debug("using memory storage")
	}
	return
}
