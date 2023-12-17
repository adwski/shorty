package main

import (
	"context"
	"errors"
	"fmt"
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
	logger, lvlParseOk := initLogger()
	defer func() {
		if errLog := logger.Sync(); errLog != nil && !errors.Is(errLog, syscall.ENOTTY) {
			log.Println("failed to sync zap logger", errLog)
		}
	}()
	if !lvlParseOk {
		defer os.Exit(1)
		return
	}

	cfg, err := config.New()
	if err != nil {
		logger.Fatal("cannot configure app", zap.Error(err))
	}

	if err = run(logger, cfg); err != nil {
		logger.Fatal("runtime error", zap.Error(err))
	}
}

func run(logger *zap.Logger, cfg *config.Shorty) error {
	var (
		ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt)
		store, err  = initStorage(ctx, logger, cfg.StorageConfig)
	)
	if err != nil {
		return fmt.Errorf("cannot configure storage: %w", err)
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
	return nil
}
