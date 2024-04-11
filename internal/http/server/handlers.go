package server

import (
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	httpmodel "github.com/adwski/shorty/internal/http/model"
	"github.com/adwski/shorty/internal/model"
	"github.com/adwski/shorty/internal/services/resolver"
	"github.com/adwski/shorty/internal/services/shortener"
	"github.com/adwski/shorty/internal/session"
	"go.uber.org/zap"
)

const (
	contentTypeJSON       = "application/json"
	contentTypePlain      = "text/plain"
	headerNameContentType = "Content-Type"

	logFieldUserID = "userID"
)

// ErrRequestCtx indicates error while getting info from request context.
var (
	ErrRequestCtx = "request context error"
)

// Ping pings the storage and returns 200 if ping is successful or 500 otherwise.
func (srv *Server) Ping(w http.ResponseWriter, r *http.Request) {
	if err := srv.statusSvc.Ping(r.Context()); err != nil {
		srv.logger.Error("ping unsuccessful", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// Stats returns storage statistics.
func (srv *Server) Stats(w http.ResponseWriter, r *http.Request) {
	reqID, ok := session.GetRequestID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		srv.logger.Error("request id was not provided in context")
		return
	}
	logf := srv.logger.With(zap.String("id", reqID))
	stats, err := srv.statusSvc.Stats(r.Context())
	logf.With(
		zap.Any("stats", stats),
		zap.Error(err),
	).Debug("stats called")

	b, err := json.Marshal(stats)
	if err != nil {
		logf.Error("cannot marshal stats response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(b); err != nil {
		logf.Error("cannot write stats body")
	}
}

// Resolve retrieves original URL of corresponding shortened URL.
func (srv *Server) Resolve(w http.ResponseWriter, r *http.Request) {
	reqID, ok := session.GetRequestID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		srv.logger.Error("request id was not provided in context")
		return
	}
	redirect, err := srv.resolverSvc.Resolve(r.Context(), r.URL.Path)
	srv.logger.With(
		zap.String("redirect", redirect),
		zap.String("id", reqID),
		zap.Error(err),
	).Debug("resolve called")
	if err != nil {
		switch {
		case errors.Is(err, resolver.ErrInvalidPath):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, model.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, model.ErrDeleted):
			w.WriteHeader(http.StatusGone)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Location", redirect)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Shorten generates short URL for provided original URL and stores it.
// Short URL is returned back.
func (srv *Server) Shorten(w http.ResponseWriter, r *http.Request) {
	u, reqID, err := session.GetUserAndReqID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		srv.logger.Error(ErrRequestCtx, zap.Error(err))
		return
	}
	logf := srv.logger.With(zap.String("id", reqID), zap.String(logFieldUserID, u.ID))

	if ct := r.Header.Get(headerNameContentType); ct != contentTypeJSON {
		w.WriteHeader(http.StatusBadRequest)
		logf.Error("incorrect Content-Type",
			zap.String("expected", contentTypeJSON),
			zap.String("got", ct))
		return
	}

	body, err := readBody(r)
	if err != nil {
		logf.Error("cannot read body", zap.Error(err))
		return
	}
	var (
		shortenReq  httpmodel.ShortenRequest
		shortenResp httpmodel.ShortenResponse
	)
	if err = json.Unmarshal(body, &shortenReq); err != nil {
		logf.Error("cannot unmarshall json", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortenResp.Result, err = srv.shortenerSvc.Shorten(r.Context(), u, shortenReq.URL)
	logf.With(
		zap.String("result", shortenResp.Result),
		zap.Error(err),
	).Debug("shorten called")
	respStatus := http.StatusCreated
	if err != nil {
		switch {
		case errors.Is(shortener.ErrInvalidURL, err),
			errors.Is(shortener.ErrUnsupportedURLScheme, err):
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(model.ErrConflict, err):
			respStatus = http.StatusConflict
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	b, err := json.Marshal(&shortenResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logf.Error("cannot marshall response", zap.Error(err))
		return
	}

	w.Header().Set(headerNameContentType, contentTypeJSON)
	w.WriteHeader(respStatus)
	if _, err = w.Write(b); err != nil {
		logf.Error("error writing json body", zap.Error(err))
	}
}

// ShortenPlain generates short URL for provided original URL and stores it.
// Short URL is returned back in plain text format.
func (srv *Server) ShortenPlain(w http.ResponseWriter, r *http.Request) {
	u, reqID, err := session.GetUserAndReqID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		srv.logger.Error(ErrRequestCtx, zap.Error(err))
		return
	}
	logf := srv.logger.With(zap.String("id", reqID), zap.String(logFieldUserID, u.ID))

	body, err := readBody(r)
	if err != nil {
		logf.Error("cannot read body", zap.Error(err))
		return
	}

	result, err := srv.shortenerSvc.Shorten(r.Context(), u, string(body))
	logf.With(
		zap.String("result", result),
		zap.Error(err),
	).Debug("shortenPlain called")
	respStatus := http.StatusCreated
	if err != nil {
		switch {
		case errors.Is(shortener.ErrInvalidURL, err),
			errors.Is(shortener.ErrUnsupportedURLScheme, err):
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(model.ErrConflict, err):
			respStatus = http.StatusConflict
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set(headerNameContentType, contentTypePlain)
	w.WriteHeader(respStatus)

	if _, err = w.Write([]byte(result)); err != nil {
		logf.Error("error writing json body", zap.Error(err))
	}
}

// GetAll retrieves all urls created by one user.
func (srv *Server) GetAll(w http.ResponseWriter, r *http.Request) {
	u, reqID, err := session.GetUserAndReqID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		srv.logger.Error(ErrRequestCtx, zap.Error(err))
		return
	}
	logf := srv.logger.With(zap.String("id", reqID), zap.String(logFieldUserID, u.ID))

	urls, err := srv.shortenerSvc.GetAll(r.Context(), u)
	logf.With(
		zap.Int("urls", len(urls)),
		zap.Error(err),
	).Debug("getURLs called")
	if err != nil {
		switch {
		case errors.Is(err, model.ErrNotFound):
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, shortener.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	b, err := json.Marshal(&urls)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logf.Error("cannot marshal url list response", zap.Error(err))
		return
	}
	w.Header().Set(headerNameContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(b); err != nil {
		logf.Error("error while writing response body", zap.Error(err))
	}
}

// ShortenBatch shortens batch of original URLs. It returns batch of short URLs
// that can be matched with originals using correlation ID.
func (srv *Server) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	u, reqID, err := session.GetUserAndReqID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		srv.logger.Error(ErrRequestCtx, zap.Error(err))
		return
	}
	logf := srv.logger.With(
		zap.String("id", reqID),
		zap.String(logFieldUserID, u.ID),
		zap.Bool("newUser", u.IsNew()))

	if ct := r.Header.Get(headerNameContentType); ct != contentTypeJSON {
		w.WriteHeader(http.StatusBadRequest)
		logf.Error("incorrect Content-Type",
			zap.String("expected", contentTypeJSON),
			zap.String("got", ct))
		return
	}

	var (
		body      []byte
		batchURLs []shortener.BatchURL
	)
	if body, err = readBody(r); err != nil {
		return
	}

	if err = json.Unmarshal(body, &batchURLs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logf.Error("cannot unmarshall urls batch from json", zap.Error(err))
		return
	}

	shortURLs, err := srv.shortenerSvc.ShortenBatch(r.Context(), u, batchURLs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logf.Error("cannot store url batch", zap.Error(err))
		return
	}

	resp, err := json.Marshal(&shortURLs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logf.Error("cannot marshall response", zap.Error(err))
		return
	}

	w.Header().Set(headerNameContentType, contentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(resp); err != nil {
		logf.Error("error writing json body", zap.Error(err))
	}
}

// DeleteBatch processes batch delete request.
// URLs are pushed to flusher queue and deleted asynchronously.
func (srv *Server) DeleteBatch(w http.ResponseWriter, r *http.Request) {
	u, reqID, err := session.GetUserAndReqID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		srv.logger.Error(ErrRequestCtx, zap.Error(err))
		return
	}
	logf := srv.logger.With(zap.String("id", reqID), zap.String(logFieldUserID, u.ID))

	if ct := r.Header.Get(headerNameContentType); ct != contentTypeJSON {
		w.WriteHeader(http.StatusBadRequest)
		logf.Error("incorrect Content-Type",
			zap.String("expected", contentTypeJSON),
			zap.String("got", ct))
		return
	}

	var (
		shorts []string
		body   []byte
	)
	if body, err = readBody(r); err != nil {
		return
	}
	if err = json.Unmarshal(body, &shorts); err != nil {
		logf.Error("cannot unmarshall json body", zap.Error(err))
		return
	}

	err = srv.shortenerSvc.DeleteBatch(r.Context(), u, shorts)
	logf.With(zap.Error(err)).Debug("DeleteBatch called")
	if err != nil {
		switch {
		case errors.Is(err, shortener.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)
		case errors.Is(err, shortener.ErrEmptyBatch):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	w.WriteHeader(http.StatusAccepted)
}

func readBody(req *http.Request) ([]byte, error) {
	defer func() { _ = req.Body.Close() }()
	r, err := getContentReader(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cannot read body: %w", err)
	}
	return b, nil
}

func getContentReader(req *http.Request) (r io.ReadCloser, err error) {
	switch req.Header.Get("Content-Encoding") {
	case "gzip":
		r, err = gzip.NewReader(req.Body)
	case "deflate":
		r, err = zlib.NewReader(req.Body)
	case "":
		r = req.Body
	default:
		err = errors.New("unknown content encoding")
	}
	return
}
