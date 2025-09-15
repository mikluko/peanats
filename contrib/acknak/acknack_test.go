package acknak_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/contrib/acknak"
	"github.com/mikluko/peanats/internal/xmock/peanatsmock"
)

func TestMiddleware(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		h := peanats.ChainMsgMiddleware(
			peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				return nil
			}),
			acknak.Middleware(),
		)
		msg := peanatsmock.NewMsgJetstream(t)

		err := h.HandleMsg(context.Background(), msg)
		require.NoError(t, err)
	})
	t.Run("ack on arrival", func(t *testing.T) {
		h := peanats.ChainMsgMiddleware(
			peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				_, err := msg.(peanats.Metadatable).Metadata()
				return err
			}),
			acknak.Middleware(acknak.MiddlewareAckPolicy(acknak.AckPolicyOnArrival)),
		)
		msg := peanatsmock.NewMsgJetstream(t)
		mock.InOrder(
			msg.EXPECT().Ack(mock.Anything).Return(nil).Once(),
			msg.EXPECT().Metadata().Return(&jetstream.MsgMetadata{}, nil).Once(),
		)
		err := h.HandleMsg(context.Background(), msg)
		require.NoError(t, err)
	})
	t.Run("ack on success", func(t *testing.T) {
		h := peanats.ChainMsgMiddleware(
			peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				_, err := msg.(peanats.Metadatable).Metadata()
				return err
			}),
			acknak.Middleware(acknak.MiddlewareAckPolicy(acknak.AckPolicyOnSuccess)),
		)
		msg := peanatsmock.NewMsgJetstream(t)
		mock.InOrder(
			msg.EXPECT().Metadata().Return(&jetstream.MsgMetadata{}, nil).Once(),
			msg.EXPECT().Ack(mock.Anything).Return(nil).Once(),
		)
		err := h.HandleMsg(context.Background(), msg)
		require.NoError(t, err)
	})
	t.Run("nak on error", func(t *testing.T) {
		handlerErr := errors.New("handler error")
		h := peanats.ChainMsgMiddleware(
			peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				return handlerErr
			}),
			acknak.Middleware(acknak.MiddlewareNakPolicy(acknak.NakPolicyOnError)),
		)
		msg := peanatsmock.NewMsgJetstream(t)
		mock.InOrder(
			msg.EXPECT().Metadata().Return(&jetstream.MsgMetadata{}, nil).Once(),
			msg.EXPECT().Nak(mock.Anything).Return(nil).Once(),
		)

		err := h.HandleMsg(context.Background(), msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, handlerErr)
	})
	t.Run("pass on ignored error", func(t *testing.T) {
		handlerErr := errors.New("handler error")
		h := peanats.ChainMsgMiddleware(
			peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				return handlerErr
			}),
			acknak.Middleware(
				acknak.MiddlewareNakPolicy(acknak.NakPolicyOnError),
				acknak.MiddlewareNakIgnore(handlerErr),
			),
		)
		msg := peanatsmock.NewMsgJetstream(t)
		msg.EXPECT().Metadata().Return(&jetstream.MsgMetadata{}, nil).Once()

		err := h.HandleMsg(context.Background(), msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, handlerErr)
	})
	t.Run("term on delivery limit", func(t *testing.T) {
		handlerErr := errors.New("handler error")
		h := peanats.ChainMsgMiddleware(
			peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				return handlerErr
			}),
			acknak.Middleware(acknak.MiddlewareDeliveryLimit(42)),
		)
		msg := peanatsmock.NewMsgJetstream(t)
		mock.InOrder(
			msg.EXPECT().Metadata().Return(&jetstream.MsgMetadata{
				NumDelivered: 42,
			}, nil).Once(),
			msg.EXPECT().TermWithReason(mock.Anything, acknak.DeliveryLimitExceeded).Return(nil).Once(),
		)
		err := h.HandleMsg(context.Background(), msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, handlerErr)
	})
	t.Run("nak with constant delay policy", func(t *testing.T) {
		handlerErr := errors.New("handler error")
		delayPolicy := &acknak.ConstantDelayPolicy{Duration: 5 * time.Second}
		h := peanats.ChainMsgMiddleware(
			peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				return handlerErr
			}),
			acknak.Middleware(
				acknak.MiddlewareNakPolicy(acknak.NakPolicyOnError),
				acknak.MiddlewareNakDelayPolicy(delayPolicy),
			),
		)
		msg := peanatsmock.NewMsgJetstream(t)
		mock.InOrder(
			msg.EXPECT().Metadata().Return(&jetstream.MsgMetadata{
				NumDelivered: 3, // Third attempt
			}, nil).Once(),
			msg.EXPECT().NackWithDelay(mock.Anything, 5*time.Second).Return(nil).Once(),
		)

		err := h.HandleMsg(context.Background(), msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, handlerErr)
	})
	t.Run("nak with linear delay policy", func(t *testing.T) {
		handlerErr := errors.New("handler error")
		delayPolicy := &acknak.LinearDelayPolicy{Base: 2 * time.Second, Max: 10 * time.Second}
		h := peanats.ChainMsgMiddleware(
			peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				return handlerErr
			}),
			acknak.Middleware(
				acknak.MiddlewareNakPolicy(acknak.NakPolicyOnError),
				acknak.MiddlewareNakDelayPolicy(delayPolicy),
			),
		)
		msg := peanatsmock.NewMsgJetstream(t)
		mock.InOrder(
			msg.EXPECT().Metadata().Return(&jetstream.MsgMetadata{
				NumDelivered: 3, // Third attempt -> 3 * 2s = 6s
			}, nil).Once(),
			msg.EXPECT().NackWithDelay(mock.Anything, 6*time.Second).Return(nil).Once(),
		)

		err := h.HandleMsg(context.Background(), msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, handlerErr)
	})
	t.Run("nak with exponential delay policy", func(t *testing.T) {
		handlerErr := errors.New("handler error")
		delayPolicy := &acknak.ExponentialDelayPolicy{Base: 1 * time.Second, Max: 30 * time.Second}
		h := peanats.ChainMsgMiddleware(
			peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				return handlerErr
			}),
			acknak.Middleware(
				acknak.MiddlewareNakPolicy(acknak.NakPolicyOnError),
				acknak.MiddlewareNakDelayPolicy(delayPolicy),
			),
		)
		msg := peanatsmock.NewMsgJetstream(t)
		mock.InOrder(
			msg.EXPECT().Metadata().Return(&jetstream.MsgMetadata{
				NumDelivered: 3, // Third attempt -> 2^(3-1) * 1s = 4s
			}, nil).Once(),
			msg.EXPECT().NackWithDelay(mock.Anything, 4*time.Second).Return(nil).Once(),
		)

		err := h.HandleMsg(context.Background(), msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, handlerErr)
	})
	t.Run("exponential delay policy with max cap", func(t *testing.T) {
		handlerErr := errors.New("handler error")
		delayPolicy := &acknak.ExponentialDelayPolicy{Base: 1 * time.Second, Max: 10 * time.Second}
		h := peanats.ChainMsgMiddleware(
			peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				return handlerErr
			}),
			acknak.Middleware(
				acknak.MiddlewareNakPolicy(acknak.NakPolicyOnError),
				acknak.MiddlewareNakDelayPolicy(delayPolicy),
			),
		)
		msg := peanatsmock.NewMsgJetstream(t)
		mock.InOrder(
			msg.EXPECT().Metadata().Return(&jetstream.MsgMetadata{
				NumDelivered: 5, // Fifth attempt -> 2^(5-1) * 1s = 16s, but capped at 10s
			}, nil).Once(),
			msg.EXPECT().NackWithDelay(mock.Anything, 10*time.Second).Return(nil).Once(),
		)

		err := h.HandleMsg(context.Background(), msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, handlerErr)
	})
}

func TestDelayPolicies(t *testing.T) {
	t.Run("ConstantDelayPolicy", func(t *testing.T) {
		policy := &acknak.ConstantDelayPolicy{Duration: 5 * time.Second}

		// Should return same delay for all attempts
		assert.Equal(t, 5*time.Second, policy.Delay(1))
		assert.Equal(t, 5*time.Second, policy.Delay(5))
		assert.Equal(t, 5*time.Second, policy.Delay(100))
	})

	t.Run("ConstantDelayPolicy with jitter", func(t *testing.T) {
		policy := &acknak.ConstantDelayPolicy{
			Duration: 10 * time.Second,
			Jitter:   0.5, // 50% jitter
		}

		// With jitter, delay should be between 5s and 10s
		for i := 0; i < 10; i++ {
			delay := policy.Delay(uint64(i + 1))
			assert.GreaterOrEqual(t, delay, 5*time.Second, "delay should be at least 5s")
			assert.LessOrEqual(t, delay, 10*time.Second, "delay should be at most 10s")
		}
	})

	t.Run("LinearDelayPolicy", func(t *testing.T) {
		policy := &acknak.LinearDelayPolicy{Base: 2 * time.Second, Max: 10 * time.Second}

		// Should increase linearly
		assert.Equal(t, 2*time.Second, policy.Delay(1))
		assert.Equal(t, 4*time.Second, policy.Delay(2))
		assert.Equal(t, 6*time.Second, policy.Delay(3))
		assert.Equal(t, 8*time.Second, policy.Delay(4))
		assert.Equal(t, 10*time.Second, policy.Delay(5))

		// Should be capped at max
		assert.Equal(t, 10*time.Second, policy.Delay(6))
		assert.Equal(t, 10*time.Second, policy.Delay(100))
	})

	t.Run("LinearDelayPolicy with jitter", func(t *testing.T) {
		policy := &acknak.LinearDelayPolicy{
			Base:   2 * time.Second,
			Max:    10 * time.Second,
			Jitter: 0.3, // 30% jitter
		}

		// Test that jitter is applied correctly
		for attempt := uint64(1); attempt <= 6; attempt++ {
			expectedBase := time.Duration(attempt) * 2 * time.Second
			if expectedBase > 10*time.Second {
				expectedBase = 10 * time.Second
			}
			minDelay := time.Duration(float64(expectedBase) * 0.7) // 30% jitter means 70-100% of base

			for i := 0; i < 5; i++ {
				delay := policy.Delay(attempt)
				assert.GreaterOrEqual(t, delay, minDelay, "delay should be at least 70%% of base for attempt %d", attempt)
				assert.LessOrEqual(t, delay, expectedBase, "delay should be at most the base for attempt %d", attempt)
			}
		}
	})

	t.Run("ExponentialDelayPolicy", func(t *testing.T) {
		policy := &acknak.ExponentialDelayPolicy{Base: 1 * time.Second, Max: 30 * time.Second}

		// Should increase exponentially
		assert.Equal(t, 1*time.Second, policy.Delay(1))  // 2^0 * 1s = 1s
		assert.Equal(t, 2*time.Second, policy.Delay(2))  // 2^1 * 1s = 2s
		assert.Equal(t, 4*time.Second, policy.Delay(3))  // 2^2 * 1s = 4s
		assert.Equal(t, 8*time.Second, policy.Delay(4))  // 2^3 * 1s = 8s
		assert.Equal(t, 16*time.Second, policy.Delay(5)) // 2^4 * 1s = 16s

		// Should be capped at max
		assert.Equal(t, 30*time.Second, policy.Delay(6))  // 2^5 * 1s = 32s > 30s
		assert.Equal(t, 30*time.Second, policy.Delay(10)) // would be much higher

		// Edge case: 0 attempt
		assert.Equal(t, time.Duration(0), policy.Delay(0))
	})

	t.Run("ExponentialDelayPolicy with jitter", func(t *testing.T) {
		policy := &acknak.ExponentialDelayPolicy{
			Base:   1 * time.Second,
			Max:    30 * time.Second,
			Jitter: 0.25, // 25% jitter
		}

		// Test that jitter is applied correctly for various attempts
		testCases := []struct {
			attempt      uint64
			expectedBase time.Duration
		}{
			{1, 1 * time.Second},   // 2^0 = 1
			{2, 2 * time.Second},   // 2^1 = 2
			{3, 4 * time.Second},   // 2^2 = 4
			{4, 8 * time.Second},   // 2^3 = 8
			{5, 16 * time.Second},  // 2^4 = 16
			{6, 30 * time.Second},  // 2^5 = 32, capped at 30
			{10, 30 * time.Second}, // would be much higher, capped at 30
		}

		for _, tc := range testCases {
			minDelay := time.Duration(float64(tc.expectedBase) * 0.75) // 25% jitter means 75-100% of base

			for i := 0; i < 5; i++ {
				delay := policy.Delay(tc.attempt)
				assert.GreaterOrEqual(t, delay, minDelay, "delay should be at least 75%% of base for attempt %d", tc.attempt)
				assert.LessOrEqual(t, delay, tc.expectedBase, "delay should be at most the base for attempt %d", tc.attempt)
			}
		}

		// Edge case: 0 attempt should always return 0
		assert.Equal(t, time.Duration(0), policy.Delay(0))
	})
}
