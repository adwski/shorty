package file

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/adwski/shorty/internal/storage/memory"
	"github.com/adwski/shorty/internal/storage/memory/db"

	"go.uber.org/zap"
)

const (
	fileBufferSize = 100000

	flushInterval = 2 * time.Second

	storageFilePermission = 0600
)

// File is a simple in-memory store with file persistence.
// Saving into file is done in background without affecting
// Get/Set operations. Since file is completely rewritten on each
// interval, this store is not suited for large quantities of records.
type File struct {
	*memory.Memory
	log      *zap.Logger
	filePath string
	changed  bool
	shutdown bool
}

type Config struct {
	Logger                 *zap.Logger
	FilePath               string
	IgnoreContentOnStartup bool
}

func New(cfg *Config) (*File, error) {
	if cfg.Logger == nil {
		return nil, errors.New("nil logger")
	}

	var (
		st  = memory.New()
		err error
	)
	if !cfg.IgnoreContentOnStartup {
		if st.DB, err = readURLsFromFile(cfg.FilePath); err != nil {
			return nil, err
		}
	}

	if ln := len(st.DB); ln > 0 {
		cfg.Logger.Info("loaded db from file",
			zap.Int("records", ln),
			zap.String("path", cfg.FilePath))
	} else {
		cfg.Logger.Info("db file empty or not exists",
			zap.String("path", cfg.FilePath))
	}

	return &File{
		Memory:   st,
		log:      cfg.Logger,
		filePath: cfg.FilePath,
	}, nil
}

func (s *File) Store(key string, url string, overwrite bool) error {
	if s.shutdown {
		return errors.New("storage is shutting down")
	}
	if err := s.Memory.Store(key, url, overwrite); err != nil {
		return fmt.Errorf("cannot store url: %w", err)
	}
	return nil
}

func (s *File) Run(ctx context.Context, wg *sync.WaitGroup) {
	s.maintainPersistence(ctx)
	wg.Done()
}

func (s *File) maintainPersistence(ctx context.Context) {
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

func (s *File) persist() {
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

func (s *File) dumpDB2File() error {
	f, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, storageFilePermission)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriterSize(f, fileBufferSize)
	defer func() { _ = w.Flush() }()

	for _, record := range s.Dump() {
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

func readURLsFromFile(filePath string) (db.DB, error) {
	f, err := os.OpenFile(filePath, syscall.O_RDONLY|syscall.O_CREAT, storageFilePermission)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReaderSize(f, fileBufferSize)

	var (
		record *db.URLRecord
		urlDB  = db.NewDB()
	)

	for {
		if record, err = readURLFromLine(r); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		urlDB[record.ShortURL] = *record
	}
	return urlDB, nil
}

func readURLFromLine(r *bufio.Reader) (*db.URLRecord, error) {
	data, err := r.ReadBytes('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) || len(data) == 0 {
			return nil, fmt.Errorf("error while reading url from filestore: %w", err)
		}
	}
	var record *db.URLRecord
	if record, err = db.NewURLRecordFromBytes(data[:len(data)-1]); err != nil {
		return nil, fmt.Errorf("cannot parse url record: %w", err)
	}
	return record, nil
}
