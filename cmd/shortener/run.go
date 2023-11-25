package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/adwski/shorty/internal/app/config"

	"github.com/adwski/shorty/internal/app"
	"go.uber.org/zap"
)

type Runnable interface {
	Run(ctx context.Context, wg *sync.WaitGroup)
}

func run(log *zap.Logger, shorty *app.Shorty, storage config.Storage) {
	var (
		wg          = &sync.WaitGroup{}
		errc        = make(chan error)
		ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt)
	)

	if st, ok := (storage).(Runnable); ok {
		wg.Add(1)
		go st.Run(ctx, wg)
	}

	wg.Add(1)
	go shorty.Run(ctx, wg, errc)

	select {
	case <-ctx.Done():
		log.Warn("app is shutting down")
	case err := <-errc:
		log.Warn("error in app", zap.Error(err))
	}
	cancel()
	wg.Wait()
}
