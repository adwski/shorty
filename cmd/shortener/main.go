package main

import (
	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/app/config"
	"github.com/sirupsen/logrus"
)

func main() {

	cfg, err := config.New()
	if err != nil {
		logrus.WithError(err).Fatal("cannot configure app")
	}

	if err = app.NewShorty(cfg).Run(); err != nil {
		cfg.Logger.WithError(err).Fatal("server failure")
	}
}
