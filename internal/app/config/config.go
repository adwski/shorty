// Package config defines configuration for Shorty app.
package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
)

// Shorty holds Shorty app config params.
type Shorty struct {
	StorageConfig   *Storage
	ListenAddr      string
	Host            string
	RedirectScheme  string
	ServedScheme    string
	JWTSecret       string
	PprofServerAddr string
	TrustRequestID  bool
}

// Storage holds Shorty storage config params.
type Storage struct {
	DatabaseDSN     string
	FileStoragePath string
	TraceDB         bool
}

// New creates Shorty config using command line argument and environment variables.
func New() (*Shorty, error) {
	var (
		listenAddr      = flag.String("a", ":8080", "listen address")
		baseURL         = flag.String("b", "http://localhost:8080", "base server URL")
		fileStoragePath = flag.String("f", "/tmp/short-url-db.json", "file storage path")
		databaseDSN     = flag.String("d", "", "postgres connection DSN")
		jwtSecret       = flag.String("jwt_secret", "supersecret", "jwt cookie secret key")
		redirectScheme  = flag.String("redirect_scheme", "", "enforce redirect scheme, leave empty to allow all")
		traceDB         = flag.Bool("trace_db", false, "print db wire protocol traces")
		trustRequestID  = flag.Bool("trust_request_id", false, "trust X-Request-Id header or generate unique requestId")
		profilerAddr    = flag.String("pprof_addr", "", "pprof server listen address, it will not start if left empty")
	)
	flag.Parse()

	//--------------------------------------------------
	// Check env vars
	//--------------------------------------------------
	envOverride("SERVER_ADDRESS", listenAddr)
	envOverride("PPROF_ADDRESS", profilerAddr)
	envOverride("BASE_URL", baseURL)
	envOverride("FILE_STORAGE_PATH", fileStoragePath)
	envOverride("DATABASE_DSN", databaseDSN)
	envOverride("JWT_SECRET", jwtSecret)

	//--------------------------------------------------
	// Parse server URL
	//--------------------------------------------------
	bURL, err := url.Parse(*baseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse base server URL: %w", err)
	}

	//--------------------------------------------------
	// Create config
	//--------------------------------------------------
	return &Shorty{
		ListenAddr:      *listenAddr,
		Host:            bURL.Host,
		RedirectScheme:  *redirectScheme,
		ServedScheme:    bURL.Scheme,
		JWTSecret:       *jwtSecret,
		TrustRequestID:  *trustRequestID,
		PprofServerAddr: *profilerAddr,
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
