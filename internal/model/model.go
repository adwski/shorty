// Package model contains common application data types.
package model

import "errors"

// Storage errors.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("conflict")
	ErrDeleted       = errors.New("deleted")
)

// URL is an url entity used by storages.
type URL struct {
	Short  string `json:"short_url"`
	Orig   string `json:"original_url"`
	UserID string `json:"-"`
	TS     int64  `json:"-"`
}

// Stats is a storage statistics.
type Stats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
