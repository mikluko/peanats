package peanats

import (
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMakeAckMiddleware(t *testing.T) {
	t.Run("pristine", func(t *testing.T) {
		inbox := nats.NewInbox()

		msgpub := new(msgPublisherMock)
		msgpub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, inbox, msg.Subject)
			require.Len(t, msg.Header, 0)
			require.Len(t, msg.Data, 0)
		}).Return(nil).Once()

		rq := new(requestMock)
		rq.On("Reply").Return(inbox)

		mw := MakeAckMiddleware(msgpub)
		h := mw(HandlerFunc(func(_ Publisher, _ Request) error {
			return nil
		}))
		err := h.Serve(&publisherMock{}, rq)
		require.NoError(t, err)
	})

	t.Run("with payload", func(t *testing.T) {
		inbox := nats.NewInbox()
		payload := []byte("the parson had a dog")

		msgpub := new(msgPublisherMock)
		msgpub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, inbox, msg.Subject)
			require.Len(t, msg.Header, 0)
			require.Equal(t, msg.Data, payload)
		}).Return(nil).Once()

		rq := new(requestMock)
		rq.On("Reply").Return(inbox)

		mw := MakeAckMiddleware(msgpub, AckMiddlewareWithPayload(payload))
		h := mw(HandlerFunc(func(_ Publisher, _ Request) error {
			return nil
		}))
		err := h.Serve(&publisherMock{}, rq)
		require.NoError(t, err)
	})

	t.Run("with headers", func(t *testing.T) {
		inbox := nats.NewInbox()
		header := nats.Header{
			"the-parson": []string{"had", "a", "dog"},
		}

		msgpub := new(msgPublisherMock)
		msgpub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, inbox, msg.Subject)
			require.Equal(t, msg.Header, header)
			require.Len(t, msg.Data, 0)
		}).Return(nil).Once()

		rq := new(requestMock)
		rq.On("Reply").Return(inbox)

		mw := MakeAckMiddleware(msgpub, AckMiddlewareWithHeader(header))
		h := mw(HandlerFunc(func(_ Publisher, _ Request) error {
			return nil
		}))
		err := h.Serve(&publisherMock{}, rq)
		require.NoError(t, err)
	})
}
