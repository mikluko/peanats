package peanats

import (
	"bytes"
	"log"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func TestAccessLogMiddleware(t *testing.T) {
	t.Run("pub", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)

		pub := publisherMock{}
		pub.On("Publish", []byte("the parson had a dog")).Return(nil).Once()
		pub.On("Subject").Return("had.a.dog").Once()
		pub.On("Header").Return(&nats.Header{})
		defer pub.AssertExpectations(t)

		req := requestMock{}
		req.On("Header").Return(nats.Header{
			HeaderRequestUID: []string{"uid:xxxxxxxxxxxx"},
		}).Once()
		req.On("Subject").Return("the.parson").Once()
		defer pub.AssertExpectations(t)

		h := ChainMiddleware(HandlerFunc(func(pub Publisher, req Request) error {
			return pub.Publish([]byte("the parson had a dog"))
		}), MakeAccessLogMiddleware(AccessLogMiddlewareWithLogger(log.New(buf, "", 0))))
		err := h.Serve(&pub, &req)
		require.NoError(t, err)

		require.Equal(t, "uid:xxxxxxxxxxxx the.parson 0 pub had.a.dog 20\n", buf.String())
	})
	t.Run("noop", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)

		pub := publisherMock{}
		defer pub.AssertExpectations(t)

		req := requestMock{}
		req.On("Header").Return(nats.Header{
			HeaderRequestUID: []string{"uid:xxxxxxxxxxxx"},
		}).Once()
		req.On("Subject").Return("the.parson.had.a.dog").Once()
		defer pub.AssertExpectations(t)

		h := ChainMiddleware(HandlerFunc(func(pub Publisher, req Request) error {
			return nil
		}), MakeAccessLogMiddleware(AccessLogMiddlewareWithLogger(log.New(buf, "", 0))))
		err := h.Serve(&pub, &req)
		require.NoError(t, err)

		require.Equal(t, "uid:xxxxxxxxxxxx the.parson.had.a.dog 0 noop\n", buf.String())
	})
}
