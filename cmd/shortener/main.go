package main

import (
	"context"
	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/app/config"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
)

func main() {

	cfg, err := config.New()
	if err != nil {
		zap.L().Fatal("cannot configure app", zap.Error(err))
	}
	log := cfg.Logger

	ctx, cancel := context.WithCancel(context.Background())
	shorty, err := app.NewShorty(ctx, cfg)
	if err != nil {
		log.Fatal("cannot create app", zap.Error(err))
	}

	var (
		wg   = &sync.WaitGroup{}
		errc = make(chan error)
	)
	wg.Add(1)
	go shorty.Run(wg, errc)

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
