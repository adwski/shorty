package buffer

import (
	"context"
	"sync"
	"time"

	"github.com/adwski/shorty/internal/storage"
	"go.uber.org/zap"
)

const (
	bufFillSize      = 10
	bufAllocSize     = 20
	bufFlushInterval = 5 * time.Second
)

type Flusher struct {
	log   *zap.Logger
	in    chan *storage.URL
	flush func(context.Context, []storage.URL)
	buf   []storage.URL
	size  int
}

func NewFlusher(log *zap.Logger, flush func(context.Context, []storage.URL)) (*Flusher, chan *storage.URL) {
	in := make(chan *storage.URL)
	return &Flusher{
		log:   log.With(zap.String("component", "flusher")),
		in:    in,
		flush: flush,
		buf:   make([]storage.URL, 0, bufAllocSize),
	}, in
}

func (s *Flusher) Run(ctx context.Context, wg *sync.WaitGroup) {
	s.log.Debug("flusher started")

	defer func() {
		s.log.Debug("flusher stopped")
		wg.Done()
	}()

	doFlush := func() {
		flushed := make([]storage.URL, len(s.buf))
		copy(flushed, s.buf)
		s.buf = s.buf[0:0]
		s.flush(ctx, flushed)
	}

	for {
		select {
		case record := <-s.in:
			s.size++
			s.buf = append(s.buf, *record)
			if len(s.buf) >= bufFillSize {
				s.log.Debug("flushing on buffer fill")
				doFlush()
			}
		case <-time.After(bufFlushInterval):
			if len(s.buf) > 0 {
				s.log.Debug("flushing on time tick")
				doFlush()
			}
		case <-ctx.Done():
			s.in = nil
			if len(s.buf) > 0 {
				s.log.Debug("flushing before shutdown")
				doFlush()
			}
			return
		}
	}
}
