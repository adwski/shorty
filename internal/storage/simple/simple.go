package simple

import (
	"fmt"
	"github.com/adwski/shorty/internal/storage/generators"
	"maps"
	"sync"

	"github.com/adwski/shorty/internal/storage/common"
)

// Simple is an in-memory URL storage
// based on map[string]string.
// All map operations are thread-safe
type Simple struct {
	st     map[string]string
	mux    *sync.Mutex
	keyLen uint
}

type Config struct {
	PathLength uint
}

func NewSimple(cfg *Config) *Simple {
	return &Simple{
		st:     make(map[string]string),
		mux:    &sync.Mutex{},
		keyLen: cfg.PathLength,
	}
}

// Get returns stored URL by specified key
func (si *Simple) Get(key string) (url string, err error) {
	var (
		ok bool
	)
	if url, ok = si.st[key]; !ok {
		err = common.ErrNotFound()
	}
	return
}

// Store stores url with specified key. If key already exists in storage
// the value will be overwritten
func (si *Simple) Store(key, url string) error {
	si.mux.Lock()
	defer si.mux.Unlock()
	si.st[key] = url
	return nil
}

// StoreUnique generates unique key for provided URL and stores it
func (si *Simple) StoreUnique(url string) (key string, err error) {
	var (
		ok bool
	)
	si.mux.Lock()
	defer si.mux.Unlock()
	for {
		key = generators.RandString(si.keyLen)
		if _, ok = si.st[key]; !ok {
			break
		}
	}
	si.st[key] = url
	return
}

func (si *Simple) Dump() string {
	si.mux.Lock()
	defer si.mux.Unlock()
	return fmt.Sprintf("%v", si.st)
}

func (si *Simple) DumpMap() map[string]string {
	si.mux.Lock()
	defer si.mux.Unlock()
	dump := make(map[string]string)
	maps.Copy(dump, si.st)
	return dump
}
