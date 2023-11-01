package main

import (
	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/app/config"
	log "github.com/sirupsen/logrus"
)

func main() {

	cfg, err := config.New()
	if err != nil {
		log.WithError(err).Fatal("cannot configure app")
	}

	log.WithFields(log.Fields{
		"address": cfg.ListenAddr,
	}).Info("starting app")

	if err = app.NewShorty(cfg).Run(); err != nil {
		log.WithError(err).Fatal("server failure")
	}
}
