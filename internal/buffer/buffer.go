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
	fillChan      chan struct{}
	flush         func(context.Context, []T)
	bufMux        *sync.Mutex
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
		flush:         flush,
		fillChan:      make(chan struct{}, 1),
		buf:           make([]T, 0, cfg.AllocSize),
		bufMux:        &sync.Mutex{},
		flushInterval: cfg.FlushInterval,
		flushSize:     cfg.FlushSize,
		shutdown:      &atomic.Bool{},
	}
}

func (s *Flusher[T]) Push(elem T) error {
	if s.shutdown.Load() {
		return errors.New("flusher is shutting down")
	}
	s.bufMux.Lock()
	defer s.bufMux.Unlock()
	s.size++
	s.buf = append(s.buf, elem)
	if len(s.buf) >= s.flushSize {
		s.fillChan <- struct{}{}
	}
	return nil
}

func (s *Flusher[T]) doFlush(ctx context.Context) {
	s.bufMux.Lock()
	defer s.bufMux.Unlock()
	if len(s.buf) > 0 {
		s.log.Debug("flushing buffer")
		flushed := make([]T, len(s.buf))
		copy(flushed, s.buf)
		s.buf = s.buf[0:0]
		s.flush(ctx, flushed)
	}
}

func (s *Flusher[T]) Run(ctx context.Context, wg *sync.WaitGroup) {
	s.log.Debug("flusher started")

	defer func() {
		s.log.Debug("flusher stopped")
		wg.Done()
	}()

	for {
		select {
		case <-s.fillChan:
		case <-time.After(s.flushInterval):
		case <-ctx.Done():
			s.shutdown.Store(true)
			s.doFlush(ctx)
			return
		}
		s.doFlush(ctx)
	}
}
