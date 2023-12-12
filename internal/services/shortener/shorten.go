package shortener

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/adwski/shorty/internal/storage"

	"github.com/adwski/shorty/internal/validate"
	"go.uber.org/zap"
)

const (
	headerContentType = "Content-Type"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
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

	if svc.redirectScheme != "" && srcURL.Scheme != svc.redirectScheme {
		w.WriteHeader(http.StatusBadRequest)
		svc.log.Error("scheme is not supported",
			zap.String("scheme", srcURL.Scheme),
			zap.String("supported", svc.redirectScheme))
		return
	}

	if shortPath, err = svc.storeURL(req.Context(), srcURL.String()); err != nil {
		if !errors.Is(err, storage.ErrConflict) {
			w.WriteHeader(http.StatusInternalServerError)
			svc.log.Error("cannot store url", zap.Error(err))
			return
		}
	}

	w.Header().Set(headerContentType, "text/plain")
	if errors.Is(err, storage.ErrConflict) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	if _, err = w.Write([]byte(svc.getServedURL(shortPath))); err != nil {
		svc.log.Error("error writing body", zap.Error(err))
	}
}

// ShortenJSON does the same as Shorten but operates with json.
func (svc *Service) ShortenJSON(w http.ResponseWriter, req *http.Request) {
	var (
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

	if svc.redirectScheme != "" && srcURL.Scheme != svc.redirectScheme {
		w.WriteHeader(http.StatusBadRequest)
		svc.log.Error("scheme is not supported",
			zap.String("scheme", srcURL.Scheme),
			zap.String("supported", svc.redirectScheme))
		return
	}

	shortPath, errStore := svc.storeURL(req.Context(), srcURL.String())
	if errStore != nil {
		if !errors.Is(errStore, storage.ErrConflict) {
			w.WriteHeader(http.StatusInternalServerError)
			svc.log.Error("cannot store url", zap.Error(err))
			return
		}
	}

	if shortenResp, err = json.Marshal(&ShortenResponse{
		Result: svc.getServedURL(shortPath),
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		svc.log.Error("cannot marshall response", zap.Error(err))
		return
	}

	w.Header().Set(headerContentType, "application/json")
	if errors.Is(errStore, storage.ErrConflict) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	if _, err = w.Write(shortenResp); err != nil {
		svc.log.Error("error writing json body", zap.Error(err))
	}
}

func getRedirectURLFromJSONBody(req *http.Request) (u *url.URL, err error) {
	var body []byte
	if body, err = readBody(req); err != nil {
		return
	}

	var shortenReq ShortenRequest
	if err = json.Unmarshal(body, &shortenReq); err != nil {
		err = fmt.Errorf("cannot unmarshall json body: %w", err)
		return
	}

	if len(shortenReq.URL) == 0 {
		err = fmt.Errorf("empty url")
		return
	}

	if u, err = url.Parse(shortenReq.URL); err != nil {
		err = fmt.Errorf("cannot parse url: %w", err)
	}
	return
}
