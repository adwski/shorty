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
	flushNeed     *atomic.Bool
	fillChan      chan struct{}
	flush         func(context.Context, []T)
	bufMux        *sync.Mutex
	buf           []T
	size          int
	allocSize     int
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
		shutdown:      &atomic.Bool{},
		flushNeed:     &atomic.Bool{},
		flushInterval: cfg.FlushInterval,
		flushSize:     cfg.FlushSize,
		allocSize:     cfg.AllocSize,
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
	if len(s.buf) >= s.flushSize && !s.flushNeed.Load() {
		s.fillChan <- struct{}{}
		s.flushNeed.Store(true)
	}
	return nil
}

func (s *Flusher[T]) doFlush(ctx context.Context) {
	s.bufMux.Lock()
	defer func() {
		s.flushNeed.Store(false)
		s.bufMux.Unlock()
	}()
	if len(s.buf) == 0 {
		return
	}
	s.log.Debug("flushing buffer", zap.Int("len", len(s.buf)))
	s.flush(ctx, s.buf)
	s.buf = make([]T, 0, s.allocSize)
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
