package file

import (
	"bufio"
	"bytes"
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

	e "github.com/adwski/shorty/internal/errors"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
)

const (
	fileReaderBufferSize = 100000

	flushInterval = 2 * time.Second

	storageFIlePermission = 0600
)

type db map[string]URLRecord

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Store struct {
	filePath string
	mux      *sync.Mutex
	db       db
	gen      uuid.Generator
	ctx      context.Context
	log      *zap.Logger
	changed  bool
	done     chan struct{}
}

type Config struct {
	FilePath string
	Ctx      context.Context
	Logger   *zap.Logger
	Done     chan struct{}
}

func New(cfg *Config) (*Store, error) {

	if cfg.Ctx == nil {
		return nil, errors.New("nil context")
	}
	if cfg.Logger == nil {
		return nil, errors.New("nil logger")
	}
	if cfg.Done == nil {
		return nil, errors.New("nil done channel")
	}

	urlDB, err := readURLsFromFile(cfg.FilePath)
	if err != nil {
		return nil, err
	}

	if ln := len(urlDB); ln > 0 {
		cfg.Logger.Info("loaded db from file",
			zap.Int("records", ln),
			zap.String("path", cfg.FilePath))
	} else {
		cfg.Logger.Info("db file empty or not exists",
			zap.String("path", cfg.FilePath))
	}

	s := &Store{
		filePath: cfg.FilePath,
		db:       urlDB,
		gen:      uuid.NewGen(),
		ctx:      cfg.Ctx,
		log:      cfg.Logger,
		done:     cfg.Done,
		mux:      &sync.Mutex{},
	}
	go s.maintainPersistence()
	return s, nil
}

func (s *Store) maintainPersistence() {
Loop:
	for {
		select {
		case <-s.ctx.Done():
			s.persist()
			s.done <- struct{}{}
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

	f, err := os.OpenFile(filePath, syscall.O_RDONLY|syscall.O_CREAT, storageFIlePermission)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReaderSize(f, fileReaderBufferSize)

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
			return nil, err
		}
	}

	var rec URLRecord
	if err = json.Unmarshal(data[:len(data)-1], &rec); err != nil {
		return nil, err
	}

	return &rec, nil
}

func (s *Store) dumpDB2File() error {
	var (
		buf bytes.Buffer
	)
	for _, record := range s.dump() {
		data, err := json.Marshal(record)
		if err != nil {
			return err
		}
		buf.Write(data)
		buf.WriteRune('\n')
	}
	return os.WriteFile(s.filePath, buf.Bytes(), storageFIlePermission)
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
		return "", e.ErrNotFound
	}
	return record.OriginalURL, nil
}

func (s *Store) Store(key string, url string, overwrite bool) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.db[key]; ok && !overwrite {
		return e.ErrAlreadyExists
	}
	u, err := s.gen.NewV4()
	if err != nil {
		return err
	}
	s.db[key] = URLRecord{
		UUID:        u.String(),
		ShortURL:    key,
		OriginalURL: url,
	}
	s.changed = true
	return nil
}
