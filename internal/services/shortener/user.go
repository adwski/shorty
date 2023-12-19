package shortener

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/adwski/shorty/internal/session"
	"github.com/adwski/shorty/internal/storage"
	"go.uber.org/zap"
)

func (svc *Service) GetURLsByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := session.GetUserFromContext(ctx)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		svc.log.Error("cannot get user from context",
			zap.Error(errors.New("middleware did not provide user context")))
		return
	}

	if user.IsNew() {
		// Session was created during this request
		// That means there is no cookie
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := svc.store.ListUserURLs(ctx, user.ID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		svc.log.Error("storage error", zap.Error(err))
		return
	}

	for i := range urls {
		urls[i].Short = svc.getServedURL(urls[i].Short)
	}

	b, errJ := json.Marshal(&urls)
	if errJ != nil {
		w.WriteHeader(http.StatusInternalServerError)
		svc.log.Error("cannot marshal url list response", zap.Error(errJ))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, errW := w.Write(b); errW != nil {
		svc.log.Error("error while writing response body", zap.Error(errW))
	}
}
