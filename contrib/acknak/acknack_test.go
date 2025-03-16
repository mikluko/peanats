package acknak_test

import (
	"context"
	"errors"
	"testing"

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
}
