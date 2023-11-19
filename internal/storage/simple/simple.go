package simple

import (
	"github.com/adwski/shorty/internal/errors"
	"sync"
)

// Store is an in-memory URL storage
// based on map[string]string.
// All map operations are thread-safe
type Store struct {
	st  map[string]string
	mux *sync.Mutex
}

func New() *Store {
	return &Store{
		st:  make(map[string]string),
		mux: &sync.Mutex{},
	}
}

// Get returns stored URL by specified key
func (s *Store) Get(key string) (url string, err error) {
	var (
		ok bool
	)
	s.mux.Lock()
	defer s.mux.Unlock()
	if url, ok = s.st[key]; !ok {
		err = errors.ErrNotFound
	}
	return
}

// Store stores url with specified key. If key already exists in storage
// the value will be overwritten
func (s *Store) Store(key, url string, overwrite bool) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.st[key]; ok && !overwrite {
		return errors.ErrAlreadyExists
	}
	s.st[key] = url
	return nil
}
