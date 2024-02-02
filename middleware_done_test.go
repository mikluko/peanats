package peanats

import (
	"errors"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDoneMiddleware(t *testing.T) {
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
			MakeDoneMiddleware(),
		)
		err := h.Serve(pub, rq)
		require.NoError(t, err)
	})

	t.Run("error from handler", func(t *testing.T) {
		inbox := nats.NewInbox()
		pub := new(publisherMock)
		handlerErr := errors.New("the parson had a dog")

		rq := new(requestMock)
		rq.On("Reply").Return(inbox)

		h := ChainMiddleware(
			HandlerFunc(func(_ Publisher, _ Request) error {
				return handlerErr
			}),
			MakeDoneMiddleware(),
		)
		err := h.Serve(pub, rq)
		require.Error(t, err)
		require.True(t, errors.Is(err, handlerErr))
	})

	t.Run("with payload", func(t *testing.T) {
		inbox := nats.NewInbox()
		payload := []byte("the parson had a dog")

		pub := new(publisherMock)
		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, inbox, msg.Subject)
			require.Equal(t, 0, len(msg.Header))
			require.Equal(t, payload, msg.Data)
		}).Return(nil).Once()

		rq := new(requestMock)
		rq.On("Reply").Return(inbox)

		h := ChainMiddleware(
			HandlerFunc(func(_ Publisher, _ Request) error {
				return nil
			}),
			MakeDoneMiddleware(DoneMiddlewareWithPayload(payload)),
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
			require.Equal(t, header, msg.Header)
			require.Len(t, msg.Data, 0)
		}).Return(nil).Once()

		rq := new(requestMock)
		rq.On("Reply").Return(inbox)

		h := ChainMiddleware(
			HandlerFunc(func(_ Publisher, _ Request) error {
				return nil
			}),
			MakeDoneMiddleware(DoneMiddlewareWithHeader(header)),
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
			MakeDoneMiddleware(DoneMiddlewareWithSubject(subject)),
		)
		err := h.Serve(pub, rq)
		require.NoError(t, err)
	})
}
