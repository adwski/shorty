package shortener

import (
	"io"
	"net/http"
	"net/url"

	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/validate"
	log "github.com/sirupsen/logrus"
)

// Service is a shortener service
type Service struct {
	store          storage.Storage
	servedScheme   string
	redirectScheme string
	host           string
}

type Config struct {
	Store          storage.Storage
	ServedScheme   string
	RedirectScheme string
	Host           string
}

func New(cfg *Config) *Service {
	return &Service{
		store:          cfg.Store,
		servedScheme:   cfg.ServedScheme,
		redirectScheme: cfg.RedirectScheme,
		host:           cfg.Host,
	}
}

// Shorten read body bytes (no more than Content-Length), parses URL from it
// and stores URL in storage. If something wrong with body or Content-Length
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
		log.WithError(err).Error("shorten request is not valid")
		return
	}

	if srcURL, err = getRedirectURLFromBody(req, bodyLength); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.WithError(err).Error("url is not valid")
		return
	}

	if svc.redirectScheme != "" && srcURL.Scheme != svc.redirectScheme {
		w.WriteHeader(http.StatusBadRequest)
		log.WithFields(log.Fields{
			"scheme":    srcURL.Scheme,
			"supported": svc.redirectScheme,
		}).Error("scheme is not supported")
		return
	}

	if shortPath, err = svc.store.StoreUnique(srcURL.String()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithError(err).Error("cannot store url")
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(svc.getServedURL(shortPath))); err != nil {
		log.WithError(err).Error("error writing body")
	}
}

func (svc *Service) getServedURL(shortPath string) string {
	return svc.servedScheme + "://" + svc.host + "/" + shortPath
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
