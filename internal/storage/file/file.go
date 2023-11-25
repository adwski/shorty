package file

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/url"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/adwski/shorty/internal/storage"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
)

const (
	fileBufferSize = 100000

	flushInterval = 2 * time.Second

	storageFilePermission = 0600
)

type db map[string]URLRecord

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Store is a simple in-memory store with file persistence.
// Saving into file is done in background without affecting
// Get/Set operations. Since file is completely rewritten on each
// interval, this store is not suited for large quantities of records.
type Store struct {
	gen      uuid.Generator
	mux      *sync.Mutex
	log      *zap.Logger
	db       db
	filePath string
	changed  bool
	shutdown bool
}

type Config struct {
	Logger                 *zap.Logger
	FilePath               string
	IgnoreContentOnStartup bool
}

func New(cfg *Config) (*Store, error) {
	if cfg.Logger == nil {
		return nil, errors.New("nil logger")
	}

	var (
		urlDB db
		err   error
	)
	if !cfg.IgnoreContentOnStartup {
		if urlDB, err = readURLsFromFile(cfg.FilePath); err != nil {
			return nil, err
		}
	} else {
		urlDB = make(db)
	}

	if ln := len(urlDB); ln > 0 {
		cfg.Logger.Info("loaded db from file",
			zap.Int("records", ln),
			zap.String("path", cfg.FilePath))
	} else {
		cfg.Logger.Info("db file empty or not exists",
			zap.String("path", cfg.FilePath))
	}

	return &Store{
		filePath: cfg.FilePath,
		db:       urlDB,
		gen:      uuid.NewGen(),
		log:      cfg.Logger,
		mux:      &sync.Mutex{},
	}, nil
}

func (s *Store) Run(ctx context.Context, wg *sync.WaitGroup) {
	s.maintainPersistence(ctx)
	wg.Done()
}

func (s *Store) maintainPersistence(ctx context.Context) {
Loop:
	for {
		select {
		case <-ctx.Done():
			s.shutdown = true
			s.persist()
			break Loop
		case <-time.After(flushInterval):
			s.persist()
		}
	}
}

func (s *Store) persist() {
	if !s.changed {
		return
	}
	if err := s.dumpDB2File(); err != nil {
		s.log.Error("cannot save db to file",
			zap.Error(err))
	} else {
		s.changed = false
		s.log.Debug("db was saved to file",
			zap.String("path", s.filePath))
	}
}

func readURLsFromFile(filePath string) (db, error) {
	f, err := os.OpenFile(filePath, syscall.O_RDONLY|syscall.O_CREAT, storageFilePermission)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReaderSize(f, fileBufferSize)

	var (
		record *URLRecord
		urlDB  = make(db)
	)

	for {
		record, err = readURLFromLine(r)
		if errors.Is(err, io.EOF) {
			break
		}
		if _, err = uuid.FromString(record.UUID); err != nil {
			return nil, fmt.Errorf("malformed uuid: %w", err)
		}
		if _, err = url.Parse(record.OriginalURL); err != nil {
			return nil, fmt.Errorf("malformed url for %s: %w", record.UUID, err)
		}
		urlDB[record.ShortURL] = *record
	}
	return urlDB, nil
}

func readURLFromLine(r *bufio.Reader) (*URLRecord, error) {
	data, err := r.ReadBytes('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) || len(data) == 0 {
			return nil, fmt.Errorf("error while reading url from filestore: %w", err)
		}
	}

	var rec URLRecord
	if err = json.Unmarshal(data[:len(data)-1], &rec); err != nil {
		return nil, fmt.Errorf("malformed json data: %w", err)
	}
	return &rec, nil
}

func (s *Store) dumpDB2File() error {
	f, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, storageFilePermission)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriterSize(f, fileBufferSize)
	defer func() { _ = w.Flush() }()

	for _, record := range s.dump() {
		var data []byte
		if data, err = json.Marshal(record); err != nil {
			return fmt.Errorf("cannot marshal to json: %w", err)
		}
		if _, err = w.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("cannot write to file: %w", err)
		}
	}
	return nil
}

func (s *Store) dump() db {
	s.mux.Lock()
	defer s.mux.Unlock()
	dump := make(db, len(s.db))
	maps.Copy(dump, s.db)
	return dump
}

func (s *Store) Get(key string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	record, ok := s.db[key]
	if !ok {
		return "", storage.ErrNotFound
	}
	return record.OriginalURL, nil
}

func (s *Store) Store(key string, url string, overwrite bool) error {
	if s.shutdown {
		return errors.New("storage is shutting down")
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.db[key]; ok && !overwrite {
		return storage.ErrAlreadyExists
	}
	u, err := s.gen.NewV4()
	if err != nil {
		return fmt.Errorf("cannot generate key uuid: %w", err)
	}
	s.db[key] = URLRecord{
		UUID:        u.String(),
		ShortURL:    key,
		OriginalURL: url,
	}
	s.changed = true
	return nil
}
