package app

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/adwski/shorty/internal/app/config"
	"github.com/adwski/shorty/internal/storage/database"
	"go.uber.org/zap"
)

func Example() { //nolint:testableexamples // no output here
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	// Init zap logger
	logger, _ := zap.NewDevelopment()

	// Init PostgreSQL storage
	dbStore, _ := database.New(ctx, &database.Config{
		Logger: logger,
		DSN:    "postgres://postgres@localhos:5432/postgres",
	})

	// Create shortener app
	shorty, _ := NewShorty(logger, dbStore, &config.Config{
		ListenAddr:     ":8080",
		ServedHost:     "localhost",
		ServedScheme:   "http",
		JWTSecret:      "superSecret",
		TrustRequestID: true,
	})

	// Run app
	wg := &sync.WaitGroup{}
	errc := make(chan error)
	go shorty.Run(ctx, wg, errc)

	cancel()
}
