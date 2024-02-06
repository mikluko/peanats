package stream

import (
	"github.com/mikluko/peanats"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestMiddleware(t *testing.T) {
	t.Run("done in middleware", func(t *testing.T) {
		h := peanats.ChainMiddleware(
			peanats.HandlerFunc(func(pub peanats.Publisher, rq peanats.Request) error {
				return pub.Publish(nil)
			}),
			Middleware,
		)
		rq := requestMock{}
		rq.On("Reply").Return("_INBOX.1")

		pub := publisherMock{}

		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, "_INBOX.1", msg.Subject)
			require.NotEmpty(t, msg.Header.Get(HeaderStreamUID))
		}).Return(nil).Once()

		pub.On("WithSubject", mock.Anything).Return(&pub).Once()

		hdr := nats.Header{}
		pub.On("Header").Return(&hdr).Twice()

		pub.On("Publish", []byte(nil)).Run(func(args mock.Arguments) {
			require.NotEmpty(t, hdr.Get(HeaderStreamUID))
			require.NotEmpty(t, hdr.Get(HeaderStreamSequence))
		}).Return(nil).Once()

		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.True(t, strings.HasPrefix(msg.Subject, StreamPrefix))
			require.NotEmpty(t, msg.Header.Get(HeaderStreamUID))
			require.Equal(t, StreamControlDone, msg.Header.Get(HeaderStreamControl))
		}).Return(nil).Once()

		err := h.Serve(&pub, &rq)
		require.NoError(t, err)

		rq.AssertExpectations(t)
		pub.AssertExpectations(t)
	})
	t.Run("done in handler", func(t *testing.T) {
		h := peanats.ChainMiddleware(
			peanats.HandlerFunc(func(pub peanats.Publisher, rq peanats.Request) error {
				pub.Header().Set(HeaderStreamControl, StreamControlDone)
				return pub.Publish(nil)
			}),
			Middleware,
		)
		rq := requestMock{}
		rq.On("Reply").Return("_INBOX.1")

		pub := publisherMock{}

		pub.On("PublishMsg", mock.Anything).Run(func(args mock.Arguments) {
			msg := args.Get(0).(*nats.Msg)
			require.Equal(t, "_INBOX.1", msg.Subject)
			require.NotEmpty(t, msg.Header.Get(HeaderStreamUID))
		}).Return(nil).Once()

		pub.On("WithSubject", mock.Anything).Return(&pub).Once()

		hdr := nats.Header{}
		pub.On("Header").Return(&hdr).Once()  // in handler
		pub.On("Header").Return(&hdr).Twice() // in publisher

		pub.On("Publish", []byte(nil)).Run(func(args mock.Arguments) {
			require.NotEmpty(t, hdr.Get(HeaderStreamUID))
			require.NotEmpty(t, hdr.Get(HeaderStreamSequence))
		}).Return(nil).Once()

		err := h.Serve(&pub, &rq)
		require.NoError(t, err)

		rq.AssertExpectations(t)
		pub.AssertExpectations(t)
	})
}
