package peanats

import (
	"context"
	"errors"
	"sync"
)

// Dispatcher provides async task dispatch with error collection and graceful drain.
type Dispatcher interface {
	Dispatch(func() error)
	Wait(context.Context) error
}

// DefaultDispatcher is the package-level default Dispatcher used when no custom Dispatcher is provided.
var DefaultDispatcher Dispatcher = NewDispatcher()

// NewDispatcher creates a Dispatcher that runs each task in a new goroutine,
// collects errors, and supports context-aware waiting.
func NewDispatcher() Dispatcher {
	return &dispatcherImpl{}
}

type dispatcherImpl struct {
	wg   sync.WaitGroup
	mu   sync.Mutex
	errs []error
}

func (d *dispatcherImpl) Dispatch(f func() error) {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		if err := f(); err != nil {
			d.mu.Lock()
			d.errs = append(d.errs, err)
			d.mu.Unlock()
		}
	}()
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
