package shortener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/adwski/shorty/internal/session"
	"github.com/adwski/shorty/internal/user"

	"github.com/adwski/shorty/internal/generators"
	"github.com/adwski/shorty/internal/storage"
	"github.com/adwski/shorty/internal/validate"
	"go.uber.org/zap"
)

type BatchURL struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type BatchShortened struct {
	ID    string `json:"correlation_id"`
	Short string `json:"short_url"`
}

// ShortenBatch shortens batch of urls.
func (svc *Service) ShortenBatch(w http.ResponseWriter, req *http.Request) {
	u, ok := session.GetUserFromContext(req.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		svc.log.Error(ErrNoUser.Error())
		return
	}
	logf := svc.log.With(zap.String("id", u.GetRequestID()))

	var (
		batchURLs   []BatchURL
		shortURLs   []BatchShortened
		shortenResp []byte
		err         error
	)
	if err = validate.ShortenRequestJSON(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logf.Error("shorten request is not valid", zap.Error(err))
		return
	}

	if batchURLs, err = getURLBatchFromJSONBody(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logf.Error("cannot get url batch from request body", zap.Error(err))
		return
	}

	if shortURLs, err = svc.shortenBatch(req.Context(), u, batchURLs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logf.Error("cannot store url batch", zap.Error(err))
		return
	}

	if shortenResp, err = json.Marshal(&shortURLs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logf.Error("cannot marshall response", zap.Error(err))
		return
	}

	w.Header().Set(headerContentType, "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(shortenResp); err != nil {
		logf.Error("error writing json body", zap.Error(err))
	}
}

func getURLBatchFromJSONBody(req *http.Request) (batchReq []BatchURL, err error) {
	var body []byte
	if body, err = readBody(req); err != nil {
		return
	}

	if err = json.Unmarshal(body, &batchReq); err != nil {
		err = fmt.Errorf("cannot unmarshall json body: %w", err)
		return
	}

	if len(batchReq) == 0 {
		err = fmt.Errorf("empty batch")
		return
	}

	for _, batchURL := range batchReq {
		if batchURL.URL == "" {
			err = errors.New("url cannot be empty")
			return
		}
		if batchURL.ID == "" {
			err = errors.New("id cannot be empty")
			return
		}
		if _, err = url.Parse(batchURL.URL); err != nil {
			err = fmt.Errorf("cannot parse url: %w", err)
		}
	}
	return
}

func (svc *Service) shortenBatch(ctx context.Context, u *user.User, batch []BatchURL) ([]BatchShortened, error) {
	var (
		err  error
		urls = make([]storage.URL, len(batch))
	)

	for i := range batch {
		urls[i].Short = generators.RandString(svc.pathLength)
		urls[i].Orig = batch[i].URL
		urls[i].UID = u.ID
	}

	if err = svc.store.StoreBatch(ctx, urls); err != nil {
		return nil, fmt.Errorf("cannot store batch: %w", err)
	}

	result := make([]BatchShortened, 0, len(batch))
	for i := range batch {
		result = append(result, BatchShortened{
			ID:    batch[i].ID,
			Short: svc.getServedURL(urls[i].Short),
		})
	}

	return result, nil
}
