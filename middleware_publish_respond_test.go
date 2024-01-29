package peanats

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishRespondMiddleware(t *testing.T) {
	ns := RunNats(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc, err := nats.Connect(ns.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Close()

	mw := MakePublishRespondMiddleware(NATS(nc))
	f := mw(Handler(HandlerFunc(func(pub Publisher, req Request) error {
		return pub.Publish([]byte("test"))
	})))

	req := new(requestMock)
	defer req.AssertExpectations(t)

	req.On("Reply").Return("test.reply")

	pub := new(publisherMock)
	defer pub.AssertExpectations(t)

	pub.On("Header").Return(&nats.Header{
		"test-header": []string{"test-value"},
	})
	pub.On("Publish", []byte("test")).Return(nil)

	sub, err := nc.SubscribeSync("test.reply")
	_ = sub.AutoUnsubscribe(1)
	require.NoError(t, err)

	err = f.Serve(pub, req)
	require.NoError(t, err)

	msg, err := sub.NextMsgWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("test"), msg.Data)
	assert.Equal(t, "test-value", msg.Header.Get("test-header"))
}
