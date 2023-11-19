package config

import (
	"errors"
	"flag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/url"
	"os"
)

const (
	StorageKindSimple = iota
	StorageKindFile
)

type ShortyConfig struct {
	ListenAddr      string
	Host            string
	RedirectScheme  string
	ServedScheme    string
	FileStoragePath string
	Storage         int
	GenerateReqID   bool
	Logger          *zap.Logger
}

func New() (*ShortyConfig, error) {

	var (
		listenAddr      = flag.String("a", ":8080", "listen address")
		baseURL         = flag.String("b", "http://localhost:8080", "base server URL")
		fileStoragePath = flag.String("f", "/tmp/short-url-db.json", "file storage path")
		redirectScheme  = flag.String("redirect_scheme", "", "enforce redirect scheme, leave empty to allow all")
		logLevel        = flag.String("log_level", "debug", "log level")
		trustRequestID  = flag.Bool("trust_request_id", false, "trust X-Request-Id header or generate unique requestId")
	)
	flag.Parse()

	//--------------------------------------------------
	// Check env vars
	//--------------------------------------------------
	envOverride("SERVER_ADDRESS", listenAddr)
	envOverride("BASE_URL", baseURL)
	envOverride("FILE_STORAGE_PATH", fileStoragePath)

	//--------------------------------------------------
	// Configure Logger
	//--------------------------------------------------
	lvl, err := zapcore.ParseLevel(*logLevel)
	if err != nil {
		lvl = zapcore.InfoLevel
	}
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	})
	logger := zap.New(zapcore.NewCore(encoder, os.Stdout, lvl))

	//--------------------------------------------------
	// Parse server URL
	//--------------------------------------------------
	bURL, err := url.Parse(*baseURL)
	if err != nil {
		return nil, errors.Join(errors.New("cannot parse base server URL"), err)
	}

	var storageKind int
	if *fileStoragePath == "" {
		logger.Info("using simple storage")
	} else {
		storageKind = StorageKindFile
		logger.Info("using file storage")
	}

	//--------------------------------------------------
	// Create config
	//--------------------------------------------------
	return &ShortyConfig{
		ListenAddr:      *listenAddr,
		Host:            bURL.Host,
		RedirectScheme:  *redirectScheme,
		ServedScheme:    bURL.Scheme,
		GenerateReqID:   !*trustRequestID,
		Logger:          logger,
		Storage:         storageKind,
		FileStoragePath: *fileStoragePath,
	}, nil
}

func envOverride(name string, param *string) {
	if param == nil {
		return
	}
	if val, ok := os.LookupEnv(name); ok {
		*param = val
	}
}
