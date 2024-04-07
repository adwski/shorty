package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"syscall"

	"github.com/adwski/shorty/internal/app"
	"github.com/adwski/shorty/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envLogLevel = "LOG_LEVEL"

	defaultLogLevel = zapcore.DebugLevel
)

var (
	buildVer    = "N/A"
	buildGoVer  = "N/A"
	buildTime   = "N/A"
	buildCommit = "N/A"
)

func getLogger() (*zap.Logger, error) {
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	})
	var (
		err                 error
		logLvl              = defaultLogLevel
		logLevel, envExists = os.LookupEnv(envLogLevel)
	)
	if envExists {
		if logLevel == "" {
			err = errors.New(envLogLevel + " env is defined but empty")
		} else {
			if logLvl, err = zapcore.ParseLevel(logLevel); err != nil {
				err = fmt.Errorf("cannot parse log level: %w", err)
			} else {
				logLvl = defaultLogLevel
			}
		}
	}
	logger := zap.New(zapcore.NewCore(encoder, os.Stdout, logLvl))
	return logger, err
}

func main() {
	logger, errLvl := getLogger()
	defer func() {
		if errLog := logger.Sync(); errLog != nil &&
			!errors.Is(errLog, syscall.EBADF) &&
			!errors.Is(errLog, syscall.EINVAL) &&
			!errors.Is(errLog, syscall.ENOTTY) {
			log.Println("failed to sync zap logger", errLog)
		}
	}()
	if errLvl != nil {
		logger.Error("cannot parse log level", zap.Error(errLvl))
		defer os.Exit(1)
		return
	}

	if bInfo, ok := debug.ReadBuildInfo(); ok {
		buildGoVer = bInfo.GoVersion
	}

	logger.Debug("build info",
		zap.String("version", buildVer),
		zap.String("go", buildGoVer),
		zap.String("time", buildTime),
		zap.String("commit", buildCommit))

	cfg, err := config.New(logger)
	if err != nil {
		logger.Fatal("cannot configure app", zap.Error(err))
	}

	defer os.Exit(app.Run(logger, cfg))
}
