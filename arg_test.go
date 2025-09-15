package peanats_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xmock/peanatsmock"
)

func TestArgAckableImpl(t *testing.T) {
	type testArg struct {
		Value string `json:"value"`
	}

	t.Run("Ackable methods are properly delegated", func(t *testing.T) {
		// Create a mock JetStream message that implements Ackable
		mockMsg := peanatsmock.NewMsgJetstream(t)
		mockMsg.EXPECT().Subject().Return("test.subject").Maybe()
		mockMsg.EXPECT().Header().Return(peanats.Header{}).Maybe()
		mockMsg.EXPECT().Data().Return([]byte(`{"value":"test"}`)).Maybe()

		// Test that all Ackable methods are properly delegated
		ctx := context.Background()

		// Create an ArgHandler that verifies Ackable methods work
		handler := peanats.MsgHandlerFromArgHandler(peanats.ArgHandlerFunc[testArg](func(ctx context.Context, arg peanats.Arg[testArg]) error {
			// Verify the arg implements Ackable
			ackable, ok := arg.(peanats.Ackable)
			require.True(t, ok, "Arg should implement Ackable when underlying message is Ackable")

			// Test Ack
			mockMsg.EXPECT().Ack(ctx).Return(nil).Once()
			err := ackable.Ack(ctx)
			assert.NoError(t, err)

			// Test Nak
			mockMsg.EXPECT().Nak(ctx).Return(nil).Once()
			err = ackable.Nak(ctx)
			assert.NoError(t, err)

			// Test NackWithDelay
			delay := 5 * time.Second
			mockMsg.EXPECT().NackWithDelay(ctx, delay).Return(nil).Once()
			err = ackable.NackWithDelay(ctx, delay)
			assert.NoError(t, err)

			// Test Term
			mockMsg.EXPECT().Term(ctx).Return(nil).Once()
			err = ackable.Term(ctx)
			assert.NoError(t, err)

			// Test TermWithReason
			mockMsg.EXPECT().TermWithReason(ctx, "test reason").Return(nil).Once()
			err = ackable.TermWithReason(ctx, "test reason")
			assert.NoError(t, err)

			// Test InProgress
			mockMsg.EXPECT().InProgress(ctx).Return(nil).Once()
			err = ackable.InProgress(ctx)
			assert.NoError(t, err)

			return nil
		}))

		// Execute the handler
		err := handler.HandleMsg(ctx, mockMsg)
		require.NoError(t, err)
	})

	t.Run("Non-Ackable message doesn't implement Ackable", func(t *testing.T) {
		// Create a regular non-ackable message mock
		mockMsg := peanatsmock.NewMsg(t)
		mockMsg.EXPECT().Subject().Return("test.subject").Maybe()
		mockMsg.EXPECT().Header().Return(peanats.Header{}).Maybe()
		mockMsg.EXPECT().Data().Return([]byte(`{"value":"test"}`)).Maybe()

		// Create an ArgHandler that verifies Ackable is not available
		handler := peanats.MsgHandlerFromArgHandler(peanats.ArgHandlerFunc[testArg](func(ctx context.Context, arg peanats.Arg[testArg]) error {
			// Verify the arg does NOT implement Ackable
			_, ok := arg.(peanats.Ackable)
			assert.False(t, ok, "Arg should not implement Ackable when underlying message is not Ackable")
			return nil
		}))

		// Execute the handler
		err := handler.HandleMsg(context.Background(), mockMsg)
		require.NoError(t, err)
	})
}

func TestArgumentMessageHandler(t *testing.T) {
	type testArg struct {
		Value string `json:"value"`
	}
	t.Run("happy path", func(t *testing.T) {
		h := peanats.MsgHandlerFromArgHandler(peanats.ArgHandlerFunc[testArg](func(ctx context.Context, arg peanats.Arg[testArg]) error {
			assert.Equal(t, "parson.had", arg.Subject())
			assert.Equal(t, peanats.Header{"x-parson": []string{"dog"}}, arg.Header())
			assert.Equal(t, &testArg{Value: "a dog"}, arg.Value())
			return nil
		}))
		m := peanatsmock.NewMsg(t)
		m.EXPECT().Subject().Return("parson.had")
		m.EXPECT().Header().Return(peanats.Header{"x-parson": []string{"dog"}})
		m.EXPECT().Data().Return([]byte(`{"value":"a dog"}`))
		err := h.HandleMsg(t.Context(), m)
		require.NoError(t, err)
	})
	t.Run("decode error", func(t *testing.T) {
		h := peanats.MsgHandlerFromArgHandler(peanats.ArgHandlerFunc[testArg](func(ctx context.Context, _ peanats.Arg[testArg]) error {
			panic("should not be called")
		}))
		m := peanatsmock.NewMsg(t)
		m.EXPECT().Header().Return(peanats.Header{"x-parson": []string{"dog"}})
		m.EXPECT().Data().Return([]byte(`{`))
		err := h.HandleMsg(t.Context(), m)
		require.Error(t, err)
		require.ErrorIs(t, err, peanats.ErrArgumentUnmarshalFailed)
	})
	t.Run("handler error", func(t *testing.T) {
		handlerErr := fmt.Errorf("test error")
		h := peanats.MsgHandlerFromArgHandler(peanats.ArgHandlerFunc[testArg](func(ctx context.Context, _ peanats.Arg[testArg]) error {
			return handlerErr
		}))
		m := peanatsmock.NewMsg(t)
		m.EXPECT().Header().Return(peanats.Header{"x-parson": []string{"dog"}})
		m.EXPECT().Data().Return([]byte(`{}`))
		err := h.HandleMsg(t.Context(), m)
		require.Error(t, err)
		assert.ErrorIs(t, err, handlerErr)
	})
}
