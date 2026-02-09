package peanats_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
)

func TestDispatcher(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		d := peanats.NewDispatcher()
		d.Dispatch(func() error { return nil })
		d.Dispatch(func() error { return nil })
		err := d.Wait(t.Context())
		require.NoError(t, err)
	})
	t.Run("collects multiple errors", func(t *testing.T) {
		d := peanats.NewDispatcher()
		errA := errors.New("error a")
		errB := errors.New("error b")
		d.Dispatch(func() error { return errA })
		d.Dispatch(func() error { return errB })
		err := d.Wait(t.Context())
		require.Error(t, err)
		assert.ErrorIs(t, err, errA)
		assert.ErrorIs(t, err, errB)
	})
	t.Run("wait with no dispatches", func(t *testing.T) {
		d := peanats.NewDispatcher()
		err := d.Wait(t.Context())
		require.NoError(t, err)
	})
	t.Run("wait with context timeout", func(t *testing.T) {
		d := peanats.NewDispatcher()
		d.Dispatch(func() error {
			time.Sleep(time.Second)
			return nil
		})
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
		defer cancel()
		err := d.Wait(ctx)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
	t.Run("concurrent dispatches", func(t *testing.T) {
		d := peanats.NewDispatcher()
		const n = 100
		for i := range n {
			d.Dispatch(func() error {
				if i%10 == 0 {
					return errors.New("every tenth fails")
				}
				return nil
			})
		}
		err := d.Wait(t.Context())
		require.Error(t, err)
		// Count individual errors inside the joined error
		joined := err.Error()
		assert.Equal(t, n/10, len(strings.Split(joined, "\n")))
	})
	t.Run("reuse after wait", func(t *testing.T) {
		d := peanats.NewDispatcher()

		// First cycle
		errFirst := errors.New("first")
		d.Dispatch(func() error { return errFirst })
		err := d.Wait(t.Context())
		require.ErrorIs(t, err, errFirst)

		// Second cycle — previous errors must not leak through
		d.Dispatch(func() error { return nil })
		err = d.Wait(t.Context())
		require.NoError(t, err)

		// Third cycle — new errors collected independently
		errThird := errors.New("third")
		d.Dispatch(func() error { return errThird })
		err = d.Wait(t.Context())
		require.ErrorIs(t, err, errThird)
		assert.NotErrorIs(t, err, errFirst)
	})
}
