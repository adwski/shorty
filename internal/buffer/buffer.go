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
	bufFlushInterval = 10 * time.Second
)

type Flusher struct {
	log   *zap.Logger
	bMux  *sync.Mutex
	in    chan *storage.URL
	flush func(context.Context, []storage.URL)
	buf   []storage.URL
	size  int
}

func NewFlusher(log *zap.Logger, in chan *storage.URL, flush func(context.Context, []storage.URL)) *Flusher {
	return &Flusher{
		log:   log.With(zap.String("component", "flusher")),
		bMux:  &sync.Mutex{},
		in:    in,
		flush: flush,
		buf:   make([]storage.URL, 0),
	}
}

func (s *Flusher) Run(ctx context.Context, wg *sync.WaitGroup) {
	s.log.Debug("flusher started")

	defer wg.Done()
	for {
		select {
		case record := <-s.in:
			s.bMux.Lock()
			s.size++
			s.buf = append(s.buf, *record)

			// flush on buffer fill
			var flushed []storage.URL
			if len(s.buf) >= bufFillSize {
				flushed = make([]storage.URL, len(s.buf))
				copy(flushed, s.buf)
				s.buf = nil
			}
			s.bMux.Unlock()

			if flushed != nil {
				s.log.Debug("flushing on buffer fill",
					zap.Int("size", len(flushed)))
				wg.Add(1)
				go func() {
					s.flush(ctx, flushed)
					wg.Done()
				}()
			}

		case <-time.After(bufFlushInterval):
			// flush on time tick
			s.bMux.Lock()
			var flushed []storage.URL
			if len(s.buf) > 0 {
				flushed = make([]storage.URL, len(s.buf))
				copy(flushed, s.buf)
				s.buf = nil
			}
			s.bMux.Unlock()

			if flushed != nil {
				s.log.Debug("flushing on time tick",
					zap.Int("size", len(flushed)))
				wg.Add(1)
				go func() {
					s.flush(ctx, flushed)
					wg.Done()
				}()
			}

		case <-ctx.Done():
			// flush before shutdown
			s.bMux.Lock()
			var flushed []storage.URL
			if len(s.buf) > 0 {
				flushed = make([]storage.URL, len(s.buf))
				copy(flushed, s.buf)
				s.buf = nil
			}
			s.bMux.Unlock()

			if flushed != nil {
				s.log.Debug("flushing before shutdown",
					zap.Int("size", len(flushed)))
				wg.Add(1)
				go func() {
					s.flush(ctx, flushed)
					wg.Done()
				}()
			}
			return
		}
	}
}
