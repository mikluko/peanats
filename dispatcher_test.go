package peanats_test

import (
	"context"
	"errors"
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
}
