package memory

import (
	"context"
	"fmt"
	"maps"
	"sync"

	"github.com/adwski/shorty/internal/storage/memory/db"
	"github.com/gofrs/uuid/v5"

	"github.com/adwski/shorty/internal/storage"
)

// Memory is an in-memory URL storage
// based on map[string]string.
// All map operations are thread-safe.
type Memory struct {
	DB  db.DB
	mux *sync.Mutex
	gen uuid.Generator
}

func New() *Memory {
	return &Memory{
		DB:  db.NewDB(),
		mux: &sync.Mutex{},
		gen: uuid.NewGen(),
	}
}

func (m *Memory) Close() {}

func (m *Memory) Get(_ context.Context, key string) (string, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	record, ok := m.DB[key]
	if !ok {
		return "", storage.ErrNotFound
	}
	if record.Deleted {
		return "", storage.ErrDeleted
	}
	return record.OriginalURL, nil
}

func (m *Memory) Store(_ context.Context, url *storage.URL, overwrite bool) (string, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if u, ok := m.DB[url.Short]; ok && !(overwrite || u.Deleted) {
		return "", storage.ErrAlreadyExists
	}
	u, err := m.gen.NewV4()
	if err != nil {
		return "", fmt.Errorf("cannot generate key uuid: %w", err)
	}
	m.DB[url.Short] = db.Record{
		UUID:        u.String(),
		ShortURL:    url.Short,
		OriginalURL: url.Orig,
		UID:         url.UID,
	}
	return "", nil
}

func (m *Memory) ListUserURLs(_ context.Context, uid string) ([]*storage.URL, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	var urls []*storage.URL
	for _, record := range m.DB {
		if record.UID == uid {
			urls = append(urls, &storage.URL{
				Short: record.ShortURL,
				Orig:  record.OriginalURL,
				UID:   uid,
			})
		}
	}
	return urls, nil
}

func (m *Memory) DeleteUserURLs(_ context.Context, urls []storage.URL) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	for _, url := range urls {
		if record, ok := m.DB[url.Short]; ok {
			if record.UID == url.UID {
				record.Deleted = true
				m.DB[url.Short] = record
			}
		}
	}
	return nil
}

func (m *Memory) StoreBatch(_ context.Context, urls []storage.URL) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	IDs := make([]string, len(urls))
	for i, url := range urls {
		if u, ok := m.DB[url.Short]; ok && !u.Deleted {
			return storage.ErrAlreadyExists
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
			UID:         url.UID,
		}
	}
	return nil
}

func (m *Memory) Dump() db.DB {
	m.mux.Lock()
	defer m.mux.Unlock()
	dump := make(db.DB, len(m.DB))
	maps.Copy(dump, m.DB)
	return dump
}

func (m *Memory) Ping(_ context.Context) error {
	return nil
}
