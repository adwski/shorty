// Package storage holds common data types used by all storages.
package storage

import "errors"

// Storage errors.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("conflict")
	ErrDeleted       = errors.New("deleted")
)

// URL is an url entity used by in-memory storages.
type URL struct {
	Short  string `json:"short_url"`
	Orig   string `json:"original_url"`
	UserID string `json:"-"`
	TS     int64  `json:"-"`
}
