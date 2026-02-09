package peanats

import (
	"context"
	"errors"
	"sync"
)

// Dispatcher provides async task dispatch with error collection and graceful drain.
//
// Dispatch submits a task for asynchronous execution. Errors returned by the
// task function are collected internally.
//
// Wait blocks until all dispatched tasks complete or the context expires, then
// returns collected errors via errors.Join. After Wait returns, the error
// accumulator is reset and the Dispatcher may be reused for another
// dispatch/wait cycle.
//
// If the context passed to Wait expires before all tasks complete, Wait returns
// the context error immediately. A background goroutine continues to wait for
// the remaining tasks — its lifetime is bounded by task completion, not by the
// Dispatcher. In practice, tasks that respect their context will finish shortly
// after cancellation, allowing the goroutine to exit. Errors from tasks that
// complete after Wait returns are retained and available to a subsequent Wait
// call.
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
