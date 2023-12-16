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
	return record.OriginalURL, nil
}

func (m *Memory) Store(_ context.Context, url *storage.URL, overwrite bool) (string, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.DB[url.Short]; ok && !overwrite {
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
	}
	return "", nil
}

func (m *Memory) StoreBatch(_ context.Context, urls []storage.URL) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	IDs := make([]string, len(urls))
	for i, url := range urls {
		if _, ok := m.DB[url.Short]; ok {
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
