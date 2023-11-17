package shortener

import (
	"encoding/json"
	"errors"
	"github.com/adwski/shorty/internal/validate"
	"go.uber.org/zap"
	"net/http"
	"net/url"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

func (svc *Service) shorten(w http.ResponseWriter, srcURL *url.URL) (shortPath string, err error) {
	if svc.redirectScheme != "" && srcURL.Scheme != svc.redirectScheme {
		err = errors.New("scheme is not supported")
		w.WriteHeader(http.StatusBadRequest)
		svc.log.Error(err.Error(),
			zap.String("scheme", srcURL.Scheme),
			zap.String("supported", svc.redirectScheme))
		return
	}

	if shortPath, err = svc.storeURL(srcURL.String()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		svc.log.Error("cannot store url", zap.Error(err))
	}
	return
}

// ShortenPlain reads body bytes (no more than Content-Length), parses URL from it
// and stores URL in storage. If something is wrong with body or Content-Length
// it returns 400 error. Stored shortened path is sent back to client.
func (svc *Service) ShortenPlain(w http.ResponseWriter, req *http.Request) {
	var (
		shortPath string
		srcURL    *url.URL
		err       error
	)

	if srcURL, err = getRedirectURLFromBody(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		svc.log.Error("url is not valid", zap.Error(err))
		return
	}

	if shortPath, err = svc.shorten(w, srcURL); err != nil {
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(svc.getServedURL(shortPath))); err != nil {
		svc.log.Error("error writing body", zap.Error(err))
	}
}

// ShortenJSON does the same as Shorten but operates with json
func (svc *Service) ShortenJSON(w http.ResponseWriter, req *http.Request) {
	var (
		shortPath   string
		srcURL      *url.URL
		shortenResp []byte
		err         error
	)
	if err = validate.ShortenRequestJSON(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		svc.log.Error("shorten request is not valid", zap.Error(err))
		return
	}

	if srcURL, err = getRedirectURLFromJSONBody(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		svc.log.Error("cannot get url from request body", zap.Error(err))
		return
	}

	if shortPath, err = svc.shorten(w, srcURL); err != nil {
		return
	}

	if shortenResp, err = json.Marshal(&ShortenResponse{
		Result: svc.getServedURL(shortPath),
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		svc.log.Error("cannot marshall response", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(shortenResp); err != nil {
		svc.log.Error("error writing body", zap.Error(err))
	}
}

func getRedirectURLFromJSONBody(req *http.Request) (*url.URL, error) {
	body, err := readBody(req)
	if err != nil {
		return nil, err
	}
	var shortenReq ShortenRequest
	if err = json.Unmarshal(body, &shortenReq); err != nil {
		return nil, err
	}
	return url.Parse(shortenReq.URL)
}
