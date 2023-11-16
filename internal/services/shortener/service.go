package shortener

import (
	errs "errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/adwski/shorty/internal/errors"
	"github.com/adwski/shorty/internal/generators"
	"github.com/adwski/shorty/internal/storage"
	"go.uber.org/zap"
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

func getRedirectURLFromBody(req *http.Request, bodyLength int) (*url.URL, error) {
	body, err := readBody(req, bodyLength)
	if err != nil {
		return nil, err
	}
	return url.Parse(string(body))
}

func readBody(req *http.Request, bodyLength int) (body []byte, err error) {
	var (
		n, readBytes int
	)
	body = make([]byte, bodyLength)
	defer func() { _ = req.Body.Close() }()
	for {
		if n, err = req.Body.Read(body); err != nil {
			if errs.Is(err, io.EOF) {
				err = nil
				break
			}
			return
		}
		readBytes += n
		if readBytes >= bodyLength || n == 0 {
			break
		}
	}
	return
}
