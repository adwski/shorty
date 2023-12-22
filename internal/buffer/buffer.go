package buffer

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type Flusher[T any] struct {
	log           *zap.Logger
	shutdown      *atomic.Bool
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

func NewFlusher[T any](cfg *FlusherConfig, flush func(context.Context, []T)) *Flusher[T] {
	return &Flusher[T]{
		log:           cfg.Logger.With(zap.String("component", "flusher")),
		in:            make(chan T),
		flush:         flush,
		buf:           make([]T, 0, cfg.AllocSize),
		flushInterval: cfg.FlushInterval,
		flushSize:     cfg.FlushSize,
		shutdown:      &atomic.Bool{},
	}
}

func (s *Flusher[T]) Push(elem T) error {
	if s.shutdown.Load() {
		return errors.New("flusher is shutting down")
	}
	s.in <- elem
	return nil
}

func (s *Flusher[T]) doFlush(ctx context.Context) {
	flushed := make([]T, len(s.buf))
	copy(flushed, s.buf)
	s.buf = s.buf[0:0]
	s.flush(ctx, flushed)
}

func (s *Flusher[T]) Run(ctx context.Context, wg *sync.WaitGroup) {
	s.log.Debug("flusher started")

	defer func() {
		s.log.Debug("flusher stopped")
		wg.Done()
	}()

	for {
		select {
		case record := <-s.in:
			s.size++
			s.buf = append(s.buf, record)
			if len(s.buf) >= s.flushSize {
				s.log.Debug("flushing on buffer fill")
				s.doFlush(ctx)
			}
		case <-time.After(s.flushInterval):
			if len(s.buf) > 0 {
				s.log.Debug("flushing on time tick")
				s.doFlush(ctx)
			}
		case <-ctx.Done():
			s.shutdown.Store(true)
			if len(s.buf) > 0 {
				s.log.Debug("flushing before shutdown")
				s.doFlush(ctx)
			}
			return
		}
	}
}
