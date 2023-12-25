package buffer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFlusher(t *testing.T) {
	type args struct {
		flushInterval time.Duration
		flushSize     int
		allocSize     int
		elemNum       int
		workers       int
	}

	type testCase struct {
		name string
		args args
	}
	tests := []testCase{
		{
			name: "process 5000 then shutdown",
			args: args{
				flushInterval: 5 * time.Second,
				flushSize:     5,
				allocSize:     10,
				elemNum:       5000,
				workers:       1,
			},
		},
		{
			name: "process 100 async from 10 workers each",
			args: args{
				flushInterval: 10 * time.Second,
				flushSize:     5,
				allocSize:     10,
				elemNum:       100,
				workers:       10,
			},
		},
		{
			name: "process 100 async from 20 workers short interval",
			args: args{
				flushInterval: 100 * time.Millisecond,
				flushSize:     5,
				allocSize:     10,
				elemNum:       100,
				workers:       20,
			},
		},
		{
			name: "very short interval and flush size",
			args: args{
				flushInterval: 1 * time.Millisecond,
				flushSize:     1,
				allocSize:     10,
				elemNum:       100,
				workers:       20,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			var (
				wg          = &sync.WaitGroup{}
				wgw         = &sync.WaitGroup{}
				ctx, cancel = context.WithCancel(context.Background())
				faker       = gofakeit.New(time.Now().UnixMicro())
				outBuf      []string
				controlBuf  []string
			)

			// Create flush callback
			flush := func(_ context.Context, data []string) {
				outBuf = append(outBuf, data...)
			}
			// Create flusher
			flusher := NewFlusher[string](&FlusherConfig{
				Logger:        logger,
				FlushInterval: tt.args.flushInterval,
				FlushSize:     tt.args.flushSize,
				AllocSize:     tt.args.allocSize,
			}, flush)
			// Run flusher
			wg.Add(1)
			go flusher.Run(ctx, wg)

			// Write to flusher concurrently
			ctrlChan := make(chan string, 1000)
			for w := 0; w < tt.args.workers; w++ {
				wgw.Add(1)
				go func() {
					for i := 0; i < tt.args.elemNum; i++ {
						fake := faker.NounAbstract()
						ctrlChan <- fake
						require.NoError(t, flusher.Push(fake))
					}
					wgw.Done()
				}()
			}
			// Gather flushes
			wg.Add(1)
			go func() {
				for elem := range ctrlChan {
					controlBuf = append(controlBuf, elem)
				}
				wg.Done()
			}()
			// Wait and finish
			wgw.Wait()
			close(ctrlChan)
			cancel()
			wg.Wait()
			// Check that len of flushed elements is equal to len of generated elements
			require.NotEmpty(t, outBuf)
			assert.Equal(t, len(controlBuf), len(outBuf))
		})
	}
}
