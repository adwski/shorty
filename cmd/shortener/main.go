package main

import (
	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/app/config"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {

	var (
		cfg *config.ShortyConfig
		err error
	)

	if cfg, err = config.New(); err != nil {
		log.WithError(err).Fatal("cannot configure app")
	}

	log.WithFields(log.Fields{
		"address": cfg.ListenAddr,
	}).Info("starting app")

	if err = app.NewShorty(cfg).Run(); err != nil {
		log.WithError(err).Errorf("server failure")
		os.Exit(1)
	}
}
