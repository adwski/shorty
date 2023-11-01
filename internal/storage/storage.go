package storage

type Storage interface {
	Get(key string) (url string, err error)
	Store(key string, url string) error
	StoreUnique(url string) (key string, err error)
}
