package peanats

import (
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestStreamMiddleware(t *testing.T) {
	t.Run("done in middleware", func(t *testing.T) {
		h := ChainMiddleware(
			HandlerFunc(func(pub Publisher, rq Request) error {
				return pub.Publish(nil)
			}),
			StreamMiddleware,
		)
		rq := requestMock{}
		rq.On("Reply").Return("_INBOX.1")

		rq.On("Header").Return(&nats.Header{}).Twice()

		pub := publisherMock{}
		pub.On("Publish", []byte(nil)).Return(nil).Once()

		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, "_INBOX.1", msg.Subject)
			require.NotEmpty(t, msg.Header.Get(HeaderStreamUID))
		}).Return(nil).Once()

		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.True(t, strings.HasPrefix(msg.Subject, StreamPrefix))
			require.NotEmpty(t, msg.Header.Get(HeaderStreamUID))
			require.Equal(t, "1", msg.Header.Get(HeaderStreamSequence))
		}).Return(nil).Once()

		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.True(t, strings.HasPrefix(msg.Subject, StreamPrefix))
			require.NotEmpty(t, msg.Header.Get(HeaderStreamUID))
		}).Return(nil).Once()

		err := h.Serve(&pub, &rq)
		require.NoError(t, err)
	})
	t.Run("done in handler", func(t *testing.T) {
		h := ChainMiddleware(
			HandlerFunc(func(pub Publisher, rq Request) error {
				pub.Header().Set(HeaderStreamControl, StreamControlDone)
				return pub.Publish(nil)
			}),
			StreamMiddleware,
		)
		rq := requestMock{}
		rq.On("Reply").Return("_INBOX.1")

		rq.On("Header").Return(&nats.Header{}).Twice()

		pub := publisherMock{}
		pub.On("Publish", []byte(nil)).Return(nil).Once()

		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, "_INBOX.1", msg.Subject)
			require.NotEmpty(t, msg.Header.Get(HeaderStreamUID))
		}).Return(nil).Once()

		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.True(t, strings.HasPrefix(msg.Subject, StreamPrefix))
			require.NotEmpty(t, msg.Header.Get(HeaderStreamUID))
			require.Equal(t, "1", msg.Header.Get(HeaderStreamSequence))
			require.Equal(t, StreamControlDone, msg.Header.Get(HeaderStreamControl))
		}).Return(nil).Once()

		err := h.Serve(&pub, &rq)
		require.NoError(t, err)
	})
}
