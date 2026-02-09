package peanats

import (
	"context"
	"errors"
	"sync"
)

// Dispatcher provides async task dispatch with error collection and graceful drain.
//
// # Methods
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
// the context error joined with any errors already collected from completed
// tasks. A background goroutine continues to wait for the remaining tasks — its
// lifetime is bounded by task completion, not by the Dispatcher. In practice,
// tasks that respect their context will finish shortly after cancellation,
// allowing the goroutine to exit.
//
// # Expected usage
//
// The Dispatcher does not own the message source lifecycle. Callers are
// expected to stop the message source (unsubscribe, stop consumer) before
// calling Wait. The typical shutdown sequence is:
//
//  1. Cancel the context — message sources stop producing.
//  2. Call Wait with a deadline — in-flight tasks drain and errors are returned.
//
// Example:
//
//	disp := peanats.NewDispatcher()
//	ch, _ := subscriber.SubscribeChan(ctx, handler, subscriber.SubscribeDispatcher(disp))
//	sub, _ := conn.SubscribeChan(ctx, subject, ch)
//	defer sub.Unsubscribe()
//
//	<-ctx.Done()
//
//	waitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err := disp.Wait(waitCtx); err != nil {
//	    slog.Error("drain failed", "error", err)
//	}
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
		d.mu.Lock()
		err := errors.Join(append(d.errs, ctx.Err())...)
		d.errs = nil
		d.mu.Unlock()
		return err
	}
}
