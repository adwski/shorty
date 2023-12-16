package main

import (
	"context"
	"fmt"
	"os"

	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/app/config"
	"github.com/adwski/shorty/internal/storage/database"
	"github.com/adwski/shorty/internal/storage/file"
	"github.com/adwski/shorty/internal/storage/memory"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envLogLevel = "LOG_LEVEL"

	defaultLogLevel = zapcore.DebugLevel
)

func initLogger() *zap.Logger {
	var (
		err      error
		logLvl   zapcore.Level
		logLevel = os.Getenv(envLogLevel)
	)
	if logLevel != "" {
		logLvl, err = zapcore.ParseLevel(logLevel)
	}
	if logLevel == "" || err != nil {
		logLvl = defaultLogLevel
	}
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	})
	logger := zap.New(zapcore.NewCore(encoder, os.Stdout, logLvl))
	if err != nil {
		logger.Error("could not parse log level", zap.Error(err))
	}
	return logger
}

func initStorage(ctx context.Context, logger *zap.Logger, cfg *config.Storage) (store app.Storage, err error) {
	switch {
	case cfg.DatabaseDSN != "":
		if store, err = database.New(ctx, &database.Config{
			Logger:  logger,
			DSN:     cfg.DatabaseDSN,
			Migrate: true,
			Trace:   cfg.TraceDB,
		}); err != nil {
			err = fmt.Errorf("cannot initialize database storage: %w", err)
			break
		}
		logger.Debug("using DB storage")

	case cfg.FileStoragePath != "":
		if store, err = file.New(ctx, &file.Config{
			FilePath: cfg.FileStoragePath,
			Logger:   logger,
		}); err != nil {
			err = fmt.Errorf("cannot initialize file storage: %w", err)
			break
		}
		logger.Debug("using file storage")

	default:
		store = memory.New()
		logger.Debug("using memory storage")
	}
	return
}
