// Package db holds data types used in memory storage.
package db

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/adwski/shorty/internal/user"

	"github.com/gofrs/uuid/v5"
)

// DB is in-memory database of shortened URLs.
// It represented as map hash->URL
type DB map[string]Record

// NewDB create new in-memory URL database.
func NewDB() DB {
	return make(DB)
}

// Map returns hash->OrigURL representation of URL database.
func (db DB) Map() map[string]string {
	kv := make(map[string]string, len(db))
	for k, v := range db {
		kv[k] = v.OriginalURL
	}
	return kv
}

// Record is single shortened URL record.
type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user"`
	Deleted     bool   `json:"deleted"`
}

// NewURLRecordFromBytes parses json encoded byte string and creates URL record from it.
func NewURLRecordFromBytes(data []byte) (*Record, error) {
	record := &Record{}
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("malformed json data: %w", err)
	}
	if _, err := uuid.FromString(record.UUID); err != nil {
		return nil, fmt.Errorf("malformed uuid: %w", err)
	}
	if _, err := url.Parse(record.OriginalURL); err != nil {
		return nil, fmt.Errorf("malformed url for %s: %w", record.UUID, err)
	}
	if _, err := user.NewFromUserID(record.UserID); err != nil {
		return nil, fmt.Errorf("malformed user id for %s: %w", record.UserID, err)
	}
	return record, nil
}
