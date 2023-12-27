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

func TestFlusherLateAsyncFlush(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	var (
		ctx, cancel = context.WithCancel(context.Background())
		wg          = &sync.WaitGroup{}
		outBuf      = make([]string, 0)
	)

	flush := func(_ context.Context, data []string) {
		go func() {
			time.Sleep(time.Second)
			outBuf = append(outBuf, data...)
		}()
	}

	flusher := NewFlusher[string](&FlusherConfig{
		Logger:        logger,
		FlushInterval: 10 * time.Second,
		FlushSize:     10,
		AllocSize:     20,
	}, flush)
	// Run flusher
	wg.Add(1)
	go flusher.Run(ctx, wg)

	require.NoError(t, flusher.Push("aaa"))
	flusher.doFlush(ctx)
	require.NoError(t, flusher.Push("bbb"))
	time.Sleep(2 * time.Second)
	assert.Equal(t, "aaa", outBuf[0])

	cancel()
	wg.Wait()
}

func TestFlusher(t *testing.T) {
	type args struct {
		flushInterval time.Duration
		flushSize     int
		allocSize     int
		elemNum       int
		workers       int
		async         bool
		asyncSleep    time.Duration
		gatherSleep   time.Duration
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
			name: "process 100 from 10 workers each",
			args: args{
				flushInterval: 10 * time.Second,
				flushSize:     5,
				allocSize:     10,
				elemNum:       100,
				workers:       10,
			},
		},
		{
			name: "process 100 from 20 workers short interval",
			args: args{
				flushInterval: 100 * time.Millisecond,
				flushSize:     5,
				allocSize:     10,
				elemNum:       100,
				workers:       20,
			},
		},
		{
			name: "process 100 async from 20 workers",
			args: args{
				flushInterval: 100 * time.Second,
				flushSize:     10,
				allocSize:     20,
				elemNum:       100,
				workers:       20,
				async:         true,
				asyncSleep:    2 * time.Second,
				gatherSleep:   3 * time.Second,
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
				cbCh        = make(chan []string, 100000)
				ctx, cancel = context.WithCancel(context.Background())
				faker       = gofakeit.New(time.Now().UnixMicro())
				outBuf      []string
				controlBuf  []string
			)

			// Create flush callback
			flush := func(_ context.Context, data []string) {
				if !tt.args.async {
					outBuf = append(outBuf, data...)
					return
				}
				go func() {
					if tt.args.asyncSleep > 0 {
						time.Sleep(tt.args.asyncSleep)
					}
					cbCh <- data
				}()
			}
			if tt.args.async {
				wg.Add(1)
				go func() {
				Loop:
					for {
						select {
						case d := <-cbCh:
							outBuf = append(outBuf, d...)
						case <-time.After(tt.args.gatherSleep):
							break Loop
						}
					}
					wg.Done()
				}()
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
