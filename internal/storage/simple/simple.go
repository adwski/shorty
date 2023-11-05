package simple

import (
	"github.com/adwski/shorty/internal/errors"
	"sync"
)

// Simple is an in-memory URL storage
// based on map[string]string.
// All map operations are thread-safe
type Simple struct {
	st  map[string]string
	mux *sync.Mutex
}

func NewSimple() *Simple {
	return &Simple{
		st:  make(map[string]string),
		mux: &sync.Mutex{},
	}
}

// Get returns stored URL by specified key
func (si *Simple) Get(key string) (url string, err error) {
	var (
		ok bool
	)
	si.mux.Lock()
	defer si.mux.Unlock()
	if url, ok = si.st[key]; !ok {
		err = errors.ErrNotFound
	}
	return
}

// Store stores url with specified key. If key already exists in storage
// the value will be overwritten
func (si *Simple) Store(key, url string, overwrite bool) error {
	si.mux.Lock()
	defer si.mux.Unlock()
	if _, ok := si.st[key]; ok && !overwrite {
		return errors.ErrAlreadyExists
	}
	si.st[key] = url
	return nil
}
