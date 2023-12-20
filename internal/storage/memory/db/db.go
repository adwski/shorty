package db

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/adwski/shorty/internal/user"

	"github.com/gofrs/uuid/v5"
)

type DB map[string]Record

func NewDB() DB {
	return make(DB)
}

func (db DB) Map() map[string]string {
	kv := make(map[string]string, len(db))
	for k, v := range db {
		kv[k] = v.OriginalURL
	}
	return kv
}

type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UID         string `json:"user"`
	Deleted     bool   `json:"deleted"`
}

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
	if _, err := user.NewFromUID(record.UID); err != nil {
		return nil, fmt.Errorf("malformed uid for %s: %w", record.UID, err)
	}
	return record, nil
}
