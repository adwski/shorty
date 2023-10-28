package simple

import (
	"fmt"
	"github.com/adwski/shorty/internal/storage/common"
	"math/rand"
	"sync"
)

type Simple struct {
	st     map[string]string
	mux    *sync.Mutex
	length int
}

type Config struct {
	URLLength int
}

func NewSimple(cfg *Config) *Simple {
	return &Simple{
		st:     make(map[string]string),
		mux:    &sync.Mutex{},
		length: cfg.URLLength,
	}
}

func (si *Simple) Get(key string) (url string, err error) {
	var (
		ok bool
	)
	if url, ok = si.get(key); !ok {
		err = common.ErrErrorNotFound()
	}
	return
}

func (si *Simple) Store(key, url string) error {
	si.store(key, url)
	return nil
}

func (si *Simple) StoreUnique(url string) (key string, err error) {
	var (
		ok bool
	)
	for {
		key = generateRandString(si.length)
		if _, ok = si.get(key); !ok {
			break
		}
	}
	si.store(key, url)
	return
}

func (si *Simple) Dump() string {
	si.mux.Lock()
	defer si.mux.Unlock()
	return fmt.Sprintf("%v", si.st)
}

func (si *Simple) get(key string) (url string, ok bool) {
	si.mux.Lock()
	defer si.mux.Unlock()
	url, ok = si.st[key]
	return
}

func (si *Simple) store(key, url string) {
	si.mux.Lock()
	defer si.mux.Unlock()
	si.st[key] = url
}

func generateRandString(length int) string {
	var (
		i, ch int
		b     = make([]byte, length)
	)
	for i = 0; i < length; i++ {
		ch = rand.Intn(62)
		if ch < 10 {
			// [0-9]
			b[i] = byte(ch + 48)
		} else if ch < 36 {
			// [A-Z]
			b[i] = byte(ch + 55)
		} else {
			// [a-z]
			b[i] = byte(ch + 61)
		}
	}
	return string(b)
}
