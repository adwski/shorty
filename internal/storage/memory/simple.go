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

func (m *Memory) Get(_ context.Context, key string) (string, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	record, ok := m.DB[key]
	if !ok {
		return "", storage.ErrNotFound
	}
	return record.OriginalURL, nil
}

func (m *Memory) Store(_ context.Context, key string, url string, overwrite bool) (string, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.DB[key]; ok && !overwrite {
		return "", storage.ErrAlreadyExists
	}
	u, err := m.gen.NewV4()
	if err != nil {
		return "", fmt.Errorf("cannot generate key uuid: %w", err)
	}
	m.DB[key] = db.URLRecord{
		UUID:        u.String(),
		ShortURL:    key,
		OriginalURL: url,
	}
	return "", nil
}

func (m *Memory) StoreBatch(_ context.Context, keys []string, urls []string) error {
	if len(keys) != len(urls) {
		return fmt.Errorf("incorrect number of arguments: keys: %d, urls: %d", len(keys), len(urls))
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	for i := range keys {
		if _, ok := m.DB[keys[i]]; ok {
			return storage.ErrAlreadyExists
		}
	}
	for i := range keys {
		u, err := m.gen.NewV4()
		if err != nil {
			return fmt.Errorf("cannot generate key uuid: %w", err)
		}
		m.DB[keys[i]] = db.URLRecord{
			UUID:        u.String(),
			ShortURL:    keys[i],
			OriginalURL: urls[i],
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
