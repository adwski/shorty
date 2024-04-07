package shortener

import (
	"context"
	"errors"
	"time"

	"github.com/adwski/shorty/internal/generators"
	"github.com/adwski/shorty/internal/model"
	"github.com/adwski/shorty/internal/user"
	"go.uber.org/zap"
)

// BatchURL is single batch element in batch shorten request.
type BatchURL struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

// BatchShortened is single batch element in batch shorten response.
type BatchShortened struct {
	ID    string `json:"correlation_id"`
	Short string `json:"short_url"`
}

// ShortenBatch shortens batch of urls.
func (svc *Service) ShortenBatch(ctx context.Context, u *user.User, batch []BatchURL) ([]BatchShortened, error) {
	var (
		err  error
		urls = make([]model.URL, len(batch))
	)
	for i := range batch {
		urls[i].Short = generators.RandString(svc.pathLength)
		urls[i].Orig = batch[i].URL
		urls[i].UserID = u.ID
	}
	if err = svc.store.StoreBatch(ctx, urls); err != nil {
		return nil, errors.Join(ErrStorageError, err)
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

// GetAll retrieves all urls created by one user.
func (svc *Service) GetAll(ctx context.Context, u *user.User) ([]*model.URL, error) {
	if u.IsNew() {
		// Session was created during this request
		// That means there is no valid cookie
		return nil, ErrUnauthorized
	}
	urls, err := svc.store.ListUserURLs(ctx, u.ID)
	if err != nil {
		return nil, errors.Join(ErrStorageError, err)
	}
	for i := range urls {
		urls[i].Short = svc.getServedURL(urls[i].Short)
	}
	return urls, nil
}

// DeleteBatch processes batch delete request.
// URLs are pushed to flusher queue and deleted asynchronously.
func (svc *Service) DeleteBatch(_ context.Context, u *user.User, shorts []string) error {
	if u.IsNew() {
		// Session was created during this request
		// That means there is no valid cookie
		return ErrUnauthorized
	}
	if len(shorts) == 0 {
		return ErrEmptyBatch
	}
	ts := time.Now().UnixMicro()
	for _, short := range shorts {
		if err := svc.flusher.Push(model.URL{
			Short:  short,
			UserID: u.ID,
			TS:     ts,
		}); err != nil {
			return errors.Join(ErrDelete, err)
		}
	}
	return nil
}

func (svc *Service) deleteURLs(ctx context.Context, urls []model.URL) {
	affected, err := svc.store.DeleteUserURLs(ctx, urls)
	if err != nil {
		svc.log.Error("storage error during batch deletion", zap.Error(err))
		return
	}
	svc.log.Debug("batch delete completed successfully",
		zap.Int64("affected", affected))
}
