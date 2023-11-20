package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"go.uber.org/zap"

	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/app/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		zap.L().Fatal("cannot configure app", zap.Error(err))
	}
	log := cfg.Logger

	shorty, err := app.NewShorty(cfg)
	if err != nil {
		log.Fatal("cannot create app", zap.Error(err))
	}

	var (
		wg          = &sync.WaitGroup{}
		errc        = make(chan error)
		ctx, cancel = context.WithCancel(context.Background())
	)
	wg.Add(1)
	go shorty.Run(ctx, wg, errc)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case sig := <-c:
		log.Warn("got signal", zap.String("signal", sig.String()))
	case err = <-errc:
		log.Warn("error in app", zap.Error(err))
	}
	cancel()
	wg.Wait()
}
