package config

import (
	"errors"
	"flag"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
)

type ShortyConfig struct {
	ListenAddr     string
	Host           string
	RedirectScheme string
	ServedScheme   string
}

func New() (*ShortyConfig, error) {

	var (
		listenAddr     = flag.String("a", ":8080", "listen address")
		baseURL        = flag.String("b", "http://localhost:8080", "base server URL")
		redirectScheme = flag.String("redirect_scheme", "", "enforce redirect scheme, leave empty to allow all")
		logLevel       = flag.String("log_level", "debug", "log level")
	)
	flag.Parse()

	//--------------------------------------------------
	// Check env vars
	//--------------------------------------------------
	envOverride("SERVER_ADDRESS", listenAddr)
	envOverride("BASE_URL", baseURL)

	//--------------------------------------------------
	// Configure Logger
	//--------------------------------------------------
	log.SetOutput(os.Stdout)

	if lvl, err := log.ParseLevel(*logLevel); err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(lvl)
	}

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
	return &ShortyConfig{
		ListenAddr:     *listenAddr,
		Host:           bURL.Host,
		RedirectScheme: *redirectScheme,
		ServedScheme:   bURL.Scheme,
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
