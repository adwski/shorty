package config

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

var (
	defaults = map[string]string{
		"listen_addr":       ":8080",
		"grpc_listen_addr":  ":9090",
		"base_url":          "http://localhost:8080",
		"jwt_secret":        "supersecret",
		"file_storage_path": "/tmp/short-url-db.json",
	}
)

func newFromFlags() (*Config, error) {
	fs := pflag.NewFlagSet("common", pflag.ContinueOnError)

	cfg := &Config{
		TLS:     &TLS{},
		Storage: &Storage{},
		Filter:  &Filter{},
	}

	fs.StringVarP(&cfg.configFilePath, "config", "c", "", "path to config file")

	fs.StringVarP(&cfg.ListenAddr, "listen_addr", "a", defaults["listen_addr"], "listen address")
	fs.StringVarP(&cfg.GRPCListenAddr, "grpc_listen_addr", "g", defaults["grpc_listen_addr"], "grpc listen address")
	fs.StringVarP(&cfg.PprofServerAddr, "pprof_addr", "p", "",
		"pprof server listen address, it will not start if left empty")
	fs.StringVarP(&cfg.BaseURL, "base_url", "b", defaults["base_url"], "base server URL")
	fs.StringVar(&cfg.JWTSecret, "jwt_secret", defaults["jwt_secret"], "jwt cookie secret key")
	fs.StringVar(&cfg.RedirectScheme, "redirect_scheme", "", "enforce redirect scheme, leave empty to allow all")
	fs.BoolVar(&cfg.TrustRequestID, "trust_request_id", false,
		"trust X-Request-Id header, if disabled unique id will be generated for each request even if header exists")
	fs.BoolVar(&cfg.GRPCReflection, "grpc_reflection", false,
		"enables grpc reflection api, useful for testing with gui clients")

	fs.StringVarP(&cfg.Storage.FileStoragePath, "file_storage_path", "f",
		defaults["file_storage_path"], "file storage path")
	fs.StringVarP(&cfg.Storage.DatabaseDSN, "dsn", "d", "", "postgres connection DSN")
	fs.BoolVar(&cfg.Storage.TraceDB, "trace_db", false, "print db wire protocol traces")

	fs.BoolVarP(&cfg.TLS.Enable, "tls_enable", "s", false,
		"enable https, use tls_cert and tls_key args to provide certificate and key")
	fs.BoolVar(&cfg.TLS.UseSelfSigned, "self_signed", false, "generate self signed cert on startup")
	fs.StringVar(&cfg.TLS.KeyPath, "tls_key", "", "path to private key")
	fs.StringVar(&cfg.TLS.CertPath, "tls_cert", "", "path to certificate")

	fs.StringVarP(&cfg.Filter.Subnets, "trusted_subnets", "t",
		"127.0.0.0/8,::1/128", "comma separated list of trusted subnets, v4 and v6")
	fs.BoolVar(&cfg.Filter.TrustXFF, "trust_x_forwarded_for", false,
		"trust x-forwarded-for header during request filtering")
	fs.BoolVar(&cfg.Filter.TrustXRealIP, "trust_x_real_ip", false, "trust x-real-ip header during request filtering")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("cannot parse command line arguments: %w", err)
	}

	return cfg, nil
}
