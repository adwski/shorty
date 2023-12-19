package storage

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("conflict")
)

type URL struct {
	Short string `json:"short_url"`
	Orig  string `json:"original_url"`
	UID   string `json:"-"`
}
