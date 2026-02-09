package pond

import (
	"context"
	"errors"
	"sync"

	"github.com/alitto/pond/v2"

	"github.com/mikluko/peanats"
)

// Dispatcher creates a Dispatcher backed by a pond worker pool with the given max concurrency.
func Dispatcher(maxConcurrency int, opts ...pond.Option) peanats.Dispatcher {
	return DispatcherPool(pond.NewPool(maxConcurrency, opts...))
}

// DispatcherPool creates a Dispatcher backed by an existing pond.Pool.
func DispatcherPool(pool pond.Pool) peanats.Dispatcher {
	return &dispatcherImpl{pool: pool}
}

type dispatcherImpl struct {
	pool pond.Pool
	mu   sync.Mutex
	errs []error
	wg   sync.WaitGroup
}

func (d *dispatcherImpl) Dispatch(f func() error) {
	d.wg.Add(1)
	d.pool.Submit(func() {
		defer d.wg.Done()
		if err := f(); err != nil {
			d.mu.Lock()
			d.errs = append(d.errs, err)
			d.mu.Unlock()
		}
	})
}

func (d *dispatcherImpl) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		d.mu.Lock()
		err := errors.Join(d.errs...)
		d.errs = nil
		d.mu.Unlock()
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
