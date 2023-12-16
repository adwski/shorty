package config

import (
	"errors"
	"flag"
	"net/url"
	"os"
)

type Shorty struct {
	StorageConfig  *Storage
	ListenAddr     string
	Host           string
	RedirectScheme string
	ServedScheme   string
	GenerateReqID  bool
}

type Storage struct {
	DatabaseDSN     string
	FileStoragePath string
	TraceDB         bool
}

func New() (*Shorty, error) {
	var (
		listenAddr      = flag.String("a", ":8080", "listen address")
		baseURL         = flag.String("b", "http://localhost:8080", "base server URL")
		fileStoragePath = flag.String("f", "/tmp/short-url-db.json", "file storage path")
		databaseDSN     = flag.String("d", "", "postgres connection DSN")
		redirectScheme  = flag.String("redirect_scheme", "", "enforce redirect scheme, leave empty to allow all")
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
	// Parse server URL
	//--------------------------------------------------
	bURL, err := url.Parse(*baseURL)
	if err != nil {
		return nil, errors.Join(errors.New("cannot parse base server URL"), err)
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
		StorageConfig: &Storage{
			TraceDB:         *traceDB,
			DatabaseDSN:     *databaseDSN,
			FileStoragePath: *fileStoragePath,
		},
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
