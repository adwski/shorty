package buffer

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Flusher[T any] struct {
	log           *zap.Logger
	in            chan T
	flush         func(context.Context, []T)
	buf           []T
	size          int
	flushSize     int
	flushInterval time.Duration
}

type FlusherConfig struct {
	Logger        *zap.Logger
	FlushInterval time.Duration
	FlushSize     int
	AllocSize     int
}

func NewFlusher[T any](cfg *FlusherConfig, flush func(context.Context, []T)) (*Flusher[T], chan T) {
	in := make(chan T)
	return &Flusher[T]{
		log:           cfg.Logger.With(zap.String("component", "flusher")),
		in:            in,
		flush:         flush,
		buf:           make([]T, 0, cfg.AllocSize),
		flushInterval: cfg.FlushInterval,
		flushSize:     cfg.FlushSize,
	}, in
}

func (s *Flusher[T]) Run(ctx context.Context, wg *sync.WaitGroup) {
	s.log.Debug("flusher started")

	defer func() {
		s.log.Debug("flusher stopped")
		wg.Done()
	}()

	doFlush := func() {
		flushed := make([]T, len(s.buf))
		copy(flushed, s.buf)
		s.buf = s.buf[0:0]
		s.flush(ctx, flushed)
	}

	for {
		select {
		case record := <-s.in:
			s.size++
			s.buf = append(s.buf, record)
			if len(s.buf) >= s.flushSize {
				s.log.Debug("flushing on buffer fill")
				doFlush()
			}
		case <-time.After(s.flushInterval):
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
