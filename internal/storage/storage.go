package storage

import "github.com/adwski/shorty/internal/storage/simple"

type Storage interface {
	Get(key string) (url string, err error)
	Store(key string, url string, overwrite bool) error
}

func NewStorageSimple() Storage {
	return simple.NewSimple()
}
