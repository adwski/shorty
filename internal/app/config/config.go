package config

import (
	"errors"
	"flag"
	"net/url"
	"os"

	"github.com/sirupsen/logrus"
)

type ShortyConfig struct {
	ListenAddr     string
	Host           string
	RedirectScheme string
	ServedScheme   string
	Logger         *logrus.Logger
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
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	if lvl, err := logrus.ParseLevel(*logLevel); err != nil {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(lvl)
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
		Logger:         logger,
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
