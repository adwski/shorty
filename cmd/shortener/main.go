package main

import (
	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/app/config"
	"go.uber.org/zap"
)

func main() {

	cfg, err := config.New()
	if err != nil {
		zap.L().Fatal("cannot configure app", zap.Error(err))
	}

	if err = app.NewShorty(cfg).Run(); err != nil {
		cfg.Logger.Fatal("server failure", zap.Error(err))
	}
}
