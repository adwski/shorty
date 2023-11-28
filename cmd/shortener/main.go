package main

import (
	"go.uber.org/zap"

	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/app/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		zap.L().Fatal("cannot configure app", zap.Error(err))
	}

	shorty := app.NewShorty(cfg)
	if err != nil {
		cfg.Logger.Fatal("cannot create app", zap.Error(err))
	}

	run(cfg.Logger, shorty, cfg.Storage)
}
