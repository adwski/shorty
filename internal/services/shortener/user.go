package shortener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/adwski/shorty/internal/session"
	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/validate"
	"go.uber.org/zap"
)

type DeleteResponse struct {
	Deleted int64 `json:"deleted"`
}

func (svc *Service) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	u, ok := session.GetUserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		svc.log.Error(ErrNoUser.Error())
		return
	}
	logf := svc.log.With(zap.String("id", u.GetRequestID()))

	if u.IsNew() {
		logf.Debug("unauthorized delete call")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := validate.ShortenRequestJSON(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logf.Error("delete request is not valid", zap.Error(err))
		return
	}

	shorts, err := getShortsFromJSONBody(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logf.Error("cannot unmarshall delete request", zap.Error(err))
		return
	}

	ts := time.Now().UnixMicro()
	for _, short := range shorts {
		svc.delURLs <- storage.URL{
			Short:  short,
			UserID: u.ID,
			TS:     ts,
		}
		svc.log.Debug("sending url for deletion",
			zap.String("short", short),
			zap.String("userID", u.ID),
			zap.Int64("ts", ts))
	}
	w.WriteHeader(http.StatusAccepted)
}

func (svc *Service) deleteURLs(ctx context.Context, urls []storage.URL) {
	affected, err := svc.store.DeleteUserURLs(ctx, urls)
	if err != nil {
		svc.log.Error("storage error during batch deletion", zap.Error(err))
		return
	}
	svc.log.Debug("batch delete completed successfully",
		zap.Int64("affected", affected))
}

func (svc *Service) GetURLs(w http.ResponseWriter, r *http.Request) {
	u, ok := session.GetUserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		svc.log.Error(ErrNoUser.Error())
		return
	}
	logf := svc.log.With(zap.String("id", u.GetRequestID()))

	if u.IsNew() {
		// Session was created during this request
		// That means there is no cookie
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := svc.store.ListUserURLs(r.Context(), u.ID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		logf.Error("storage error", zap.Error(err))
		return
	}

	for i := range urls {
		urls[i].Short = svc.getServedURL(urls[i].Short)
	}

	b, errJ := json.Marshal(&urls)
	if errJ != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logf.Error("cannot marshal url list response", zap.Error(errJ))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, errW := w.Write(b); errW != nil {
		logf.Error("error while writing response body", zap.Error(errW))
	}
}

func getShortsFromJSONBody(r *http.Request) (shorts []string, err error) {
	var body []byte
	if body, err = readBody(r); err != nil {
		return
	}

	if err = json.Unmarshal(body, &shorts); err != nil {
		err = fmt.Errorf("cannot unmarshall json body: %w", err)
		return
	}

	if len(shorts) == 0 {
		err = fmt.Errorf("empty short urls list")
		return
	}
	return
}
