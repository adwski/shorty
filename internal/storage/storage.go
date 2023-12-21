package storage

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("conflict")
	ErrDeleted       = errors.New("deleted")
)

type URL struct {
	Short  string `json:"short_url"`
	Orig   string `json:"original_url"`
	UserID string `json:"-"`
}
