package simple

import (
	"fmt"
	"github.com/adwski/shorty/internal/storage/generators"
	"maps"
	"sync"

	"github.com/adwski/shorty/internal/storage/common"
)

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

func (si *Simple) Get(key string) (url string, err error) {
	var (
		ok bool
	)
	if url, ok = si.st[key]; !ok {
		err = common.ErrNotFound()
	}
	return
}

func (si *Simple) Store(key, url string) error {
	si.mux.Lock()
	defer si.mux.Unlock()
	si.st[key] = url
	return nil
}

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
