package peanats

import (
	"errors"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
	"testing"
)

func TestErrorHandlerMiddleware(t *testing.T) {

	handleErr := errors.New("handle error")
	publishErr := errors.New("publish error")
	ackErr := errors.New("ack error")

	t.Run("pass through", func(t *testing.T) {
		rq := new(requestMock)
		defer rq.AssertExpectations(t)

		pub := new(publisherMock)
		defer pub.AssertExpectations(t)

		var f Handler
		f = HandlerFunc(func(pub Publisher, req Request) error { return nil })
		f = ChainMiddleware(f, ErrorHandlerMiddleware)

		err := f.Serve(pub, rq)
		require.NoError(t, err)
	})

	t.Run("publish", func(t *testing.T) {
		rq := new(requestMock)
		defer rq.AssertExpectations(t)

		pub := new(publisherMock)
		defer pub.AssertExpectations(t)

		var f Handler
		f = HandlerFunc(func(pub Publisher, req Request) error { return handleErr })
		f = ChainMiddleware(f, ErrorHandlerMiddleware)

		header := make(nats.Header)
		pub.On("Header").Return(&header)

		pub.On("Publish", []byte(nil)).Return(nil)

		err := f.Serve(pub, rq)
		require.NoError(t, err)

		require.Equal(t, "500", header.Get(HeaderErrorCode))
		require.Equal(t, handleErr.Error(), header.Get(HeaderErrorMessage))
	})

	t.Run("publish if acked", func(t *testing.T) {
		rq := new(requestMock)
		defer rq.AssertExpectations(t)

		pub := new(publisherAckerMock)
		defer pub.AssertExpectations(t)

		var f Handler
		f = HandlerFunc(func(pub Publisher, req Request) error {
			ack := pub.(Acker)
			_ = ack.Ack(nil)
			return handleErr
		})
		f = ChainMiddleware(f, ErrorHandlerMiddleware)

		header := make(nats.Header)
		pub.On("Header").Return(&header)
		pub.On("Ack", []byte(nil)).Return(nil)
		pub.On("Publish", []byte(nil)).Return(nil)

		err := f.Serve(pub, rq)
		require.NoError(t, err)

		require.Equal(t, "500", header.Get(HeaderErrorCode))
		require.Equal(t, handleErr.Error(), header.Get(HeaderErrorMessage))
	})

	t.Run("error if publish failed", func(t *testing.T) {
		rq := new(requestMock)
		defer rq.AssertExpectations(t)

		pub := new(publisherMock)
		defer pub.AssertExpectations(t)

		var f Handler
		f = HandlerFunc(func(pub Publisher, req Request) error { return handleErr })
		f = ChainMiddleware(f, ErrorHandlerMiddleware)

		header := make(nats.Header)
		pub.On("Header").Return(&header)
		pub.On("Publish", []byte(nil)).Return(publishErr)

		err := f.Serve(pub, rq)
		require.Error(t, err)
		require.Contains(t, multierr.Errors(err), handleErr)
		require.Contains(t, multierr.Errors(err), publishErr)
	})

	t.Run("ack", func(t *testing.T) {
		rq := new(requestMock)
		defer rq.AssertExpectations(t)

		pub := new(publisherAckerMock)
		defer pub.AssertExpectations(t)

		var f Handler
		f = HandlerFunc(func(pub Publisher, req Request) error { return handleErr })
		f = ChainMiddleware(f, ErrorHandlerMiddleware)

		header := make(nats.Header)
		pub.On("Header").Return(&header)
		pub.On("Ack", []byte(nil)).Return(nil)

		err := f.Serve(pub, rq)
		require.NoError(t, err)

		require.Equal(t, "500", header.Get(HeaderErrorCode))
		require.Equal(t, handleErr.Error(), header.Get(HeaderErrorMessage))
	})

	t.Run("error if ack failed", func(t *testing.T) {
		rq := new(requestMock)
		defer rq.AssertExpectations(t)

		pub := new(publisherAckerMock)
		defer pub.AssertExpectations(t)

		var f Handler
		f = HandlerFunc(func(pub Publisher, req Request) error { return handleErr })
		f = ChainMiddleware(f, ErrorHandlerMiddleware)

		header := make(nats.Header)
		pub.On("Header").Return(&header)
		pub.On("Ack", []byte(nil)).Return(ackErr)

		err := f.Serve(pub, rq)
		require.Error(t, err)
		require.Contains(t, multierr.Errors(err), handleErr)
		require.Contains(t, multierr.Errors(err), ackErr)
	})

}
