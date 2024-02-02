package peanats

import (
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAckMiddleware(t *testing.T) {
	t.Run("pristine", func(t *testing.T) {
		inbox := nats.NewInbox()

		pub := new(publisherMock)
		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, inbox, msg.Subject)
			require.Len(t, msg.Header, 0)
			require.Len(t, msg.Data, 0)
		}).Return(nil).Once()

		rq := new(requestMock)
		rq.On("Reply").Return(inbox)

		h := ChainMiddleware(
			HandlerFunc(func(_ Publisher, _ Request) error {
				return nil
			}),
			MakeAckMiddleware(),
		)
		err := h.Serve(pub, rq)
		require.NoError(t, err)
	})

	t.Run("with payload", func(t *testing.T) {
		inbox := nats.NewInbox()
		payload := []byte("the parson had a dog")

		pub := new(publisherMock)
		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, inbox, msg.Subject)
			require.Len(t, msg.Header, 0)
			require.Equal(t, msg.Data, payload)
		}).Return(nil).Once()

		rq := new(requestMock)
		rq.On("Reply").Return(inbox)

		h := ChainMiddleware(
			HandlerFunc(func(_ Publisher, _ Request) error {
				return nil
			}),
			MakeAckMiddleware(AckMiddlewareWithPayload(payload)),
		)
		err := h.Serve(pub, rq)
		require.NoError(t, err)
	})

	t.Run("with headers", func(t *testing.T) {
		inbox := nats.NewInbox()
		header := nats.Header{
			"the-parson": []string{"had", "a", "dog"},
		}

		pub := new(publisherMock)
		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, inbox, msg.Subject)
			require.Equal(t, msg.Header, header)
			require.Len(t, msg.Data, 0)
		}).Return(nil).Once()

		rq := new(requestMock)
		rq.On("Reply").Return(inbox)

		h := ChainMiddleware(
			HandlerFunc(func(_ Publisher, _ Request) error {
				return nil
			}),
			MakeAckMiddleware(AckMiddlewareWithHeader(header)),
		)
		err := h.Serve(pub, rq)
		require.NoError(t, err)
	})

	t.Run("with subject", func(t *testing.T) {
		inbox := nats.NewInbox()
		subject := "the.parson.had.a.dog"

		pub := new(publisherMock)
		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, subject, msg.Subject)
			require.Equal(t, msg.Header, nats.Header{})
			require.Len(t, msg.Data, 0)
		}).Return(nil).Once()

		rq := new(requestMock)
		rq.On("Reply").Return(inbox)

		h := ChainMiddleware(
			HandlerFunc(func(_ Publisher, _ Request) error {
				return nil
			}),
			MakeAckMiddleware(AckMiddlewareWithSubject(subject)),
		)
		err := h.Serve(pub, rq)
		require.NoError(t, err)
	})
}
