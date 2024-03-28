package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/app/config"
	"github.com/adwski/shorty/internal/profiler"
	"go.uber.org/zap"
)

var (
	buildVer    = "N/A"
	buildGoVer  = "N/A"
	buildTime   = "N/A"
	buildCommit = "N/A"
)

func main() {
	logger, errLvl := initLogger()
	defer func() {
		if errLog := logger.Sync(); errLog != nil &&
			!errors.Is(errLog, syscall.EBADF) &&
			!errors.Is(errLog, syscall.EINVAL) &&
			!errors.Is(errLog, syscall.ENOTTY) {
			log.Println("failed to sync zap logger", errLog)
		}
	}()
	if errLvl != nil {
		logger.Error("cannot parse log level", zap.Error(errLvl))
		defer os.Exit(1)
		return
	}

	if bInfo, ok := debug.ReadBuildInfo(); ok {
		buildGoVer = bInfo.GoVersion
	}

	logger.Debug("build info",
		zap.String("version", buildVer),
		zap.String("go", buildGoVer),
		zap.String("time", buildTime),
		zap.String("commit", buildCommit))

	cfg, err := config.New(logger)
	if err != nil {
		logger.Fatal("cannot configure app", zap.Error(err))
	}

	if err = run(logger, cfg); err != nil {
		logger.Fatal("runtime error", zap.Error(err))
	}
}

func run(logger *zap.Logger, cfg *config.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	store, err := initStorage(ctx, logger, cfg.Storage)
	if err != nil {
		return fmt.Errorf("cannot configure storage: %w", err)
	}

	var (
		wg     = &sync.WaitGroup{}
		errc   = make(chan error)
		shorty = app.NewShorty(logger, store, cfg)
	)

	if cfg.PprofServerAddr != "" {
		prof := profiler.New(&profiler.Config{
			Logger:        logger,
			ListenAddress: cfg.PprofServerAddr,
		})
		wg.Add(1)
		go prof.Run(ctx, wg, errc)
	}

	wg.Add(1)
	go shorty.Run(ctx, wg, errc)

	select {
	case <-ctx.Done():
		logger.Warn("shutting down")
	case <-errc:
		cancel()
	}
	wg.Wait()
	store.Close()
	return nil
}
