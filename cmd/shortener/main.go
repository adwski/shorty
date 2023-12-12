package main

import (
	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/app/config"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		zap.NewExample().Fatal("cannot configure app", zap.Error(err))
	}

	shorty := app.NewShorty(cfg)
	if err != nil {
		cfg.Logger.Fatal("cannot create app", zap.Error(err))
	}

	run(cfg.Logger, shorty, cfg.Storage)
}
