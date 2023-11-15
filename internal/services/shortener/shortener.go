package shortener

import (
	"fmt"
	"github.com/adwski/shorty/internal/errors"
	"github.com/adwski/shorty/internal/generators"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"

	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/validate"
)

const (
	defaultMaxTries = 10
)

// Service is a shortener service
type Service struct {
	store          storage.Storage
	servedScheme   string
	redirectScheme string
	host           string
	log            *zap.Logger
	pathLength     uint
}

type Config struct {
	Store          storage.Storage
	ServedScheme   string
	RedirectScheme string
	Host           string
	Logger         *zap.Logger
	PathLength     uint
}

func New(cfg *Config) *Service {
	return &Service{
		store:          cfg.Store,
		servedScheme:   cfg.ServedScheme,
		redirectScheme: cfg.RedirectScheme,
		host:           cfg.Host,
		pathLength:     cfg.PathLength,
		log:            cfg.Logger,
	}
}

// Shorten reads body bytes (no more than Content-Length), parses URL from it
// and stores URL in storage. If something is wrong with body or Content-Length
// it returns 400 error. Stored shortened path is sent back to client.
func (svc *Service) Shorten(w http.ResponseWriter, req *http.Request) {
	var (
		bodyLength int
		shortPath  string
		srcURL     *url.URL
		err        error
	)
	if bodyLength, err = validate.ShortenRequest(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		svc.log.Error("shorten request is not valid", zap.Error(err))
		return
	}

	if srcURL, err = getRedirectURLFromBody(req, bodyLength); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		svc.log.Error("url is not valid", zap.Error(err))
		return
	}

	if svc.redirectScheme != "" && srcURL.Scheme != svc.redirectScheme {
		w.WriteHeader(http.StatusBadRequest)
		svc.log.Error("scheme is not supported",
			zap.String("scheme", srcURL.Scheme),
			zap.String("supported", svc.redirectScheme))
		return
	}

	if shortPath, err = svc.storeURL(srcURL.String()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		svc.log.Error("cannot store url", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(svc.getServedURL(shortPath))); err != nil {
		svc.log.Error("error writing body", zap.Error(err))
	}
}

func (svc *Service) getServedURL(shortPath string) string {
	return fmt.Sprintf("%s://%s/%s", svc.servedScheme, svc.host, shortPath)
}

func (svc *Service) storeURL(u string) (path string, err error) {
	for try := 0; try <= defaultMaxTries; try++ {
		if try == defaultMaxTries {
			err = errors.ErrGiveUP
			return
		}
		path = generators.RandString(svc.pathLength)
		if err = svc.store.Store(path, u, false); err != nil {
			if errors.Equal(err, errors.ErrAlreadyExists) {
				continue
			}
			return
		}
		break
	}
	return
}

func getRedirectURLFromBody(req *http.Request, bodyLength int) (u *url.URL, err error) {
	var (
		n, readBytes int
		body         = make([]byte, bodyLength)
	)
	defer func() { _ = req.Body.Close() }()
	for {
		if n, err = req.Body.Read(body); err != nil {
			if err != io.EOF {
				return
			}
			break
		}
		readBytes += n
		if readBytes >= bodyLength || n == 0 {
			break
		}
	}
	u, err = url.Parse(string(body))
	return
}
