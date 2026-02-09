package pond

import (
	"context"
	"errors"
	"fmt"
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
	if f == nil {
		return
	}
	d.wg.Add(1)
	if err := d.pool.Go(func() {
		defer d.wg.Done()
		if err := f(); err != nil {
			d.mu.Lock()
			d.errs = append(d.errs, err)
			d.mu.Unlock()
		}
	}); err != nil {
		d.wg.Done()
		d.mu.Lock()
		d.errs = append(d.errs, fmt.Errorf("dispatch failed: %w", err))
		d.mu.Unlock()
	}
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
		d.mu.Lock()
		err := errors.Join(append(d.errs, ctx.Err())...)
		d.errs = nil
		d.mu.Unlock()
		return err
	}
}
