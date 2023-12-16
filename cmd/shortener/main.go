package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/app/config"
	"go.uber.org/zap"
)

func main() {
	logger := initLogger()
	defer func() {
		if errLog := logger.Sync(); errLog != nil && !errors.Is(errLog, syscall.ENOTTY) {
			log.Println("failed to sync zap logger", errLog)
		}
	}()

	cfg, err := config.New()
	if err != nil {
		logger.Fatal("cannot configure app", zap.Error(err))
	}

	run(logger, cfg)
}

func run(logger *zap.Logger, cfg *config.Shorty) {
	var (
		ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt)
		store, errS = initStorage(ctx, logger, cfg.StorageConfig)
	)
	if errS != nil {
		logger.Fatal("cannot configure storage", zap.Error(errS))
	}

	var (
		wg     = &sync.WaitGroup{}
		errc   = make(chan error)
		shorty = app.NewShorty(logger, store, cfg)
	)
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
}
