// Package memory is simple in-memory shortened URL model.
package memory

import (
	"context"
	"fmt"
	"maps"
	"sync"

	"github.com/adwski/shorty/internal/model"
	"github.com/adwski/shorty/internal/storage/memory/db"
	"github.com/gofrs/uuid/v5"
)

// Memory is an in-memory URL storage
// based on map[string]string.
// All map operations are thread-safe.
type Memory struct {
	DB  db.DB
	mux *sync.Mutex
	gen uuid.Generator
}

// New create new memory model.
func New() *Memory {
	return &Memory{
		DB:  db.NewDB(),
		mux: &sync.Mutex{},
		gen: uuid.NewGen(),
	}
}

// Close does nothing. It's here just to comply to shortener interface.
func (m *Memory) Close() {}

// Get retrieves URL from model.
func (m *Memory) Get(_ context.Context, key string) (string, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	record, ok := m.DB[key]
	if !ok {
		return "", model.ErrNotFound
	}
	if record.Deleted {
		return "", model.ErrDeleted
	}
	return record.OriginalURL, nil
}

// Store stores shortened URL in model.
func (m *Memory) Store(_ context.Context, url *model.URL, overwrite bool) (string, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if u, ok := m.DB[url.Short]; ok && !overwrite && !u.Deleted {
		return "", model.ErrAlreadyExists
	}
	u, err := m.gen.NewV4()
	if err != nil {
		return "", fmt.Errorf("cannot generate key uuid: %w", err)
	}
	m.DB[url.Short] = db.Record{
		UUID:        u.String(),
		ShortURL:    url.Short,
		OriginalURL: url.Orig,
		UserID:      url.UserID,
	}
	return "", nil
}

// ListUserURLs returns all URL by specified user.
func (m *Memory) ListUserURLs(_ context.Context, userID string) ([]*model.URL, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	var urls []*model.URL
	for _, record := range m.DB {
		if record.UserID == userID {
			urls = append(urls, &model.URL{
				Short:  record.ShortURL,
				Orig:   record.OriginalURL,
				UserID: userID,
			})
		}
	}
	return urls, nil
}

// DeleteUserURLs deleted batch of URLs.
func (m *Memory) DeleteUserURLs(_ context.Context, urls []model.URL) (int64, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	var num int64
	for _, url := range urls {
		if record, ok := m.DB[url.Short]; ok {
			if record.UserID == url.UserID {
				record.Deleted = true
				m.DB[url.Short] = record
				num++
			}
		}
	}
	return num, nil
}

// StoreBatch stores URL batch.
func (m *Memory) StoreBatch(_ context.Context, urls []model.URL) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	IDs := make([]string, len(urls))
	for i, url := range urls {
		if u, ok := m.DB[url.Short]; ok && !u.Deleted {
			return model.ErrAlreadyExists
		}
		u, err := m.gen.NewV4()
		if err != nil {
			return fmt.Errorf("cannot generate key uuid: %w", err)
		}
		IDs[i] = u.String()
	}
	for i, url := range urls {
		m.DB[url.Short] = db.Record{
			UUID:        IDs[i],
			ShortURL:    url.Short,
			OriginalURL: url.Orig,
			UserID:      url.UserID,
		}
	}
	return nil
}

// Dump returns copy of in-memory URL database.
func (m *Memory) Dump() db.DB {
	m.mux.Lock()
	defer m.mux.Unlock()
	dump := make(db.DB, len(m.DB))
	maps.Copy(dump, m.DB)
	return dump
}

// Ping does nothing. It's here just to comply to shortener interface.
func (m *Memory) Ping(_ context.Context) error {
	return nil
}

// Stats returns total number of unique urls and users.
func (m *Memory) Stats(_ context.Context) (*model.Stats, error) {
	return m.DB.Stats(), nil
}
