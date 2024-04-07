// Package config defines configuration for Shorty app.
package config

import (
	"crypto/tls"
	"fmt"
	"net/url"

	authorizer "github.com/adwski/shorty/internal/auth"
	"github.com/adwski/shorty/internal/filter"
	"go.uber.org/zap"
)

// Config holds Shorty app config params.
type Config struct {
	Storage *Storage `json:"storage"`
	TLS     *TLS     `json:"tls"`
	Filter  *Filter  `json:"filter"`

	tls *tls.Config

	auth *authorizer.Auth

	filter *filter.Filter

	configFilePath string

	ListenAddr      string `json:"listen_addr"`
	GRPCListenAddr  string `json:"grpc_listen_addr"`
	BaseURL         string `json:"base_url"`
	RedirectScheme  string `json:"redirect_scheme"`
	JWTSecret       string `json:"jwt_secret"`
	PprofServerAddr string `json:"pprof_listen_addr"`
	ServedHost      string `json:"-"`
	ServedScheme    string `json:"-"`

	TrustRequestID bool `json:"trust_request_id"`
}

// GetTLSConfig returns crypto/tls.Config if tls was enabled in configuration,
// otherwise it will return nil.
func (cfg *Config) GetTLSConfig() *tls.Config {
	return cfg.tls
}

// GetAuthorizer retrieves application's authorizer instance.
func (cfg *Config) GetAuthorizer() *authorizer.Auth {
	return cfg.auth
}

// GetFilter retrieves application's filter instance.
func (cfg *Config) GetFilter() *filter.Filter {
	return cfg.filter
}

// TLS holds Shorty tls configuration params.
type TLS struct {
	CertPath      string `json:"cert"`
	KeyPath       string `json:"key"`
	Enable        bool   `json:"enable"`
	UseSelfSigned bool   `json:"self_signed"`
}

// Filter holds ip filter config params.
type Filter struct {
	Subnets      string `json:"trusted_subnets"`
	TrustXFF     bool   `json:"trust_x_forwarded_for"`
	TrustXRealIP bool   `json:"trust_x_real_ip"`
}

// Storage holds Shorty storage config params.
type Storage struct {
	DatabaseDSN     string `json:"database_dsn"`
	FileStoragePath string `json:"file_storage_path"`
	TraceDB         bool   `json:"trace_db"`
}

// New creates Shorty config using config file, command line argument
// and environment variables in mentioned order.
func New(logger *zap.Logger) (*Config, error) {
	var cfg *Config

	// Read flags
	cfgFromArgs, err := newFromFlags()
	if err != nil {
		return nil, err
	}

	envOverride("CONFIG", &cfgFromArgs.configFilePath)

	// Read config file
	if cfgFromArgs.configFilePath != "" {
		cfgFromFile, errCfg := newFromFile(cfgFromArgs.configFilePath)
		if errCfg != nil {
			return nil, errCfg
		}
		// Merge configs
		merge(cfgFromFile, cfgFromArgs)
		cfg = cfgFromFile
	} else {
		cfg = cfgFromArgs
	}

	// Merge Envs
	if err = mergeEnvs(cfg); err != nil {
		return nil, err
	}

	// Parse server base URL
	if err = cfg.parseBaseURL(); err != nil {
		return nil, err
	}

	if cfg.TLS.Enable {
		// Create TLS Config.
		// We must call it after base URL is parsed.
		if err = cfg.createTLSConfig(logger); err != nil {
			return nil, fmt.Errorf("tls config error: %w", err)
		}
	}

	cfg.auth = authorizer.New(cfg.JWTSecret)

	cfg.filter, err = filter.New(&filter.Config{
		Logger:             logger,
		Subnets:            cfg.Filter.Subnets,
		TrustXForwardedFor: cfg.Filter.TrustXFF,
		TrustXRealIP:       cfg.Filter.TrustXRealIP,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot configure filter: %w", err)
	}

	return cfg, nil
}

func (cfg *Config) parseBaseURL() error {
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return fmt.Errorf("cannot parse base server URL: %w", err)
	}
	cfg.ServedHost = baseURL.Host
	cfg.ServedScheme = baseURL.Scheme
	return nil
}

func (cfg *Config) createTLSConfig(logger *zap.Logger) error {
	var err error
	cfg.tls, err = getTLSConfig(logger, cfg.TLS, cfg.ServedHost)
	return err
}
