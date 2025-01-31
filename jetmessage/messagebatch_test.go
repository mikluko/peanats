package jetmessage

import (
	"context"
	"io"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/testutil"
)

type testMessageType struct {
	Seq uint `json:"seq"`
}

func TestTypedMessageBatch(t *testing.T) {
	ns := testutil.NatsServer(t)
	nc := testutil.NatsConn(t, ns)

	js, err := jetstream.New(nc)
	require.NoError(t, err)

	stream, err := js.CreateStream(context.TODO(), jetstream.StreamConfig{
		Name:     "test-stream",
		Subjects: []string{"test"},
	})
	require.NoError(t, err)

	pub := peanats.NewTypedPublisher[testMessageType](peanats.JsonCodec{}, peanats.NewPublisher(nc)).WithSubject("test")
	for i := 0; i < 10; i++ {
		err = pub.Publish(&testMessageType{Seq: uint(i)})
		require.NoError(t, err)
	}

	cons, err := stream.CreateConsumer(context.TODO(), jetstream.ConsumerConfig{
		Durable: "test-consumer",
	})
	require.NoError(t, err)

	raw, err := cons.Fetch(10)
	require.NoError(t, err)

	batch := NewMessageBatch[testMessageType](peanats.JsonCodec{}, raw)

	for i := 0; i < 10; i++ {
		msg, err := batch.Next(context.Background())
		require.NoError(t, err)
		require.Equal(t, uint(i), msg.Payload().Seq)
	}

	_, err = batch.Next(context.Background())
	require.ErrorIs(t, err, io.EOF)
}
