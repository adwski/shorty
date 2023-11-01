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

func New() (cfg *ShortyConfig, err error) {

	var (
		lvl  log.Level
		bURL *url.URL

		listenAddr     = flag.String("a", ":8080", "listen address")
		baseURL        = flag.String("b", "http://localhost:8080", "base server URL")
		redirectScheme = flag.String("redirect_scheme", "", "enforce redirect scheme, leave empty to allow all")
		logLevel       = flag.String("log_level", "debug", "log level")
	)

	flag.Parse()
	envOverride("SERVER_ADDRESS", listenAddr)
	envOverride("BASE_URL", baseURL)

	//--------------------------------------------------
	// Configure Logger
	//--------------------------------------------------
	log.SetOutput(os.Stdout)

	if lvl, err = log.ParseLevel(*logLevel); err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(lvl)
	}

	//--------------------------------------------------
	// Parse server URL
	//--------------------------------------------------
	if bURL, err = url.Parse(*baseURL); err != nil {
		err = errors.Join(errors.New("cannot parse base server URL"), err)
	}

	//--------------------------------------------------
	// Create config
	//--------------------------------------------------
	cfg = &ShortyConfig{
		ListenAddr:     *listenAddr,
		Host:           bURL.Host,
		RedirectScheme: *redirectScheme,
		ServedScheme:   bURL.Scheme,
	}
	return
}

func envOverride(name string, oldVal *string) {
	if oldVal == nil {
		return
	}
	if val, ok := os.LookupEnv(name); ok {
		*oldVal = val
	}
	return
}
