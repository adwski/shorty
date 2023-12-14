package config

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/adwski/shorty/internal/storage/postgres"

	"net/url"
	"os"

	"github.com/adwski/shorty/internal/storage/file"
	"github.com/adwski/shorty/internal/storage/memory"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Storage interface {
	Get(ctx context.Context, key string) (url string, err error)
	Store(ctx context.Context, key string, url string, overwrite bool) (string, error)
	StoreBatch(ctx context.Context, keys []string, urls []string) error
	Ping(ctx context.Context) error
}

type Shorty struct {
	Storage        Storage
	Logger         *zap.Logger
	ListenAddr     string
	Host           string
	RedirectScheme string
	ServedScheme   string
	GenerateReqID  bool
}

func New() (*Shorty, error) {
	var (
		listenAddr      = flag.String("a", ":8080", "listen address")
		baseURL         = flag.String("b", "http://localhost:8080", "base server URL")
		fileStoragePath = flag.String("f", "/tmp/short-url-db.json", "file storage path")
		databaseDSN     = flag.String("d", "", "postgres connection DSN")
		redirectScheme  = flag.String("redirect_scheme", "", "enforce redirect scheme, leave empty to allow all")
		logLevel        = flag.String("log_level", "debug", "log level")
		traceDB         = flag.Bool("trace_db", true, "print db wire protocol traces")
		trustRequestID  = flag.Bool("trust_request_id", false, "trust X-Request-Id header or generate unique requestId")
	)
	flag.Parse()

	//--------------------------------------------------
	// Check env vars
	//--------------------------------------------------
	envOverride("SERVER_ADDRESS", listenAddr)
	envOverride("BASE_URL", baseURL)
	envOverride("FILE_STORAGE_PATH", fileStoragePath)
	envOverride("DATABASE_DSN", databaseDSN)

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
		TimeKey:        "time",
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

	//--------------------------------------------------
	// Init storage
	//--------------------------------------------------
	storage, err := initStorage(logger, databaseDSN, fileStoragePath, traceDB)
	if err != nil {
		return nil, err
	}

	//--------------------------------------------------
	// Create config
	//--------------------------------------------------
	return &Shorty{
		ListenAddr:     *listenAddr,
		Host:           bURL.Host,
		RedirectScheme: *redirectScheme,
		ServedScheme:   bURL.Scheme,
		GenerateReqID:  !*trustRequestID,
		Logger:         logger,
		Storage:        storage,
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

func initStorage(
	logger *zap.Logger,
	databaseDSN, fileStoragePath *string,
	traceDB *bool,
) (storage Storage, err error) {
	switch {
	case *databaseDSN != "":
		if storage, err = postgres.New(&postgres.Config{
			Logger:  logger,
			DSN:     *databaseDSN,
			Migrate: true,
			Trace:   *traceDB,
		}); err != nil {
			return nil, fmt.Errorf("cannot initialize postgres storage: %w", err)
		}
		logger.Info("using postgres storage")

	case *fileStoragePath != "":
		if storage, err = file.New(&file.Config{
			FilePath: *fileStoragePath,
			Logger:   logger,
		}); err != nil {
			return nil, fmt.Errorf("cannot initialize file storage: %w", err)
		}
		logger.Info("using file storage")

	default:
		storage = memory.New()
		logger.Info("using memory storage")
	}
	return
}
