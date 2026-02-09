package pond_test

import (
	"context"
	"errors"
	"testing"
	"time"

	pondlib "github.com/alitto/pond/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats/contrib/pond"
)

func TestDispatcher(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		d := pond.Dispatcher(4)
		d.Dispatch(func() error { return nil })
		d.Dispatch(func() error { return nil })
		err := d.Wait(t.Context())
		require.NoError(t, err)
	})
	t.Run("collects errors", func(t *testing.T) {
		d := pond.Dispatcher(4)
		errA := errors.New("error a")
		errB := errors.New("error b")
		d.Dispatch(func() error { return errA })
		d.Dispatch(func() error { return errB })
		err := d.Wait(t.Context())
		require.Error(t, err)
		assert.ErrorIs(t, err, errA)
		assert.ErrorIs(t, err, errB)
	})
	t.Run("wait with context timeout", func(t *testing.T) {
		d := pond.Dispatcher(4)
		d.Dispatch(func() error {
			time.Sleep(time.Second)
			return nil
		})
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
		defer cancel()
		err := d.Wait(ctx)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
	t.Run("dispatch after pool stopped", func(t *testing.T) {
		pool := pondlib.NewPool(4)
		d := pond.DispatcherPool(pool)

		// Stop the pool before dispatching
		pool.StopAndWait()

		// Dispatch to stopped pool must not hang Wait
		d.Dispatch(func() error { return nil })

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		err := d.Wait(ctx)
		require.Error(t, err)
		assert.ErrorIs(t, err, pondlib.ErrPoolStopped)
	})
}
