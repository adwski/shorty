package main

import (
	"flag"
	"os"

	"github.com/adwski/shorty/internal/app"
	log "github.com/sirupsen/logrus"
)

func main() {

	var (
		err    error
		lvl    log.Level
		shorty *app.Shorty

		listenAddr     = flag.String("listen", ":8080", "listen address")
		scheme         = flag.String("scheme", "http", "server scheme")
		host           = flag.String("host", "localhost:8080", "server host")
		redirectScheme = flag.String("redirect_scheme", "", "enforce redirect scheme, leave empty to allow all")
		logLevel       = flag.String("log_level", "debug", "log level")
	)

	//--------------------------------------------------
	// Configure Logger
	//--------------------------------------------------
	log.SetOutput(os.Stdout)

	if lvl, err = log.ParseLevel(*logLevel); err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(lvl)
	}

	//--------------------------------------------------
	// Spawn app
	//--------------------------------------------------
	shorty = app.NewShorty(&app.ShortyConfig{
		ListenAddr:     *listenAddr,
		Host:           *host,
		RedirectScheme: *redirectScheme,
		ServedScheme:   *scheme,
	})

	log.WithFields(log.Fields{
		"address": *listenAddr,
	}).Info("starting app")

	if err = shorty.Run(); err != nil {
		log.WithError(err).Errorf("server failure")
		os.Exit(1)
	}
}
