package jetconsumer

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/jetmessage"
	"github.com/mikluko/peanats/testutil"
)

type testMessageType struct {
	Seq uint `json:"seq"`
}

func TestTypedBatchConsumer(t *testing.T) {
	ns := testutil.NatsServer(t)
	nc := testutil.NatsConn(t, ns)

	js, err := jetstream.New(nc)
	require.NoError(t, err)

	stream, err := js.CreateStream(context.TODO(), jetstream.StreamConfig{
		Name:     t.Name(),
		Subjects: []string{t.Name()},
	})
	require.NoError(t, err)
	consumer, err := stream.CreateConsumer(context.TODO(), jetstream.ConsumerConfig{
		Durable:   t.Name(),
		AckPolicy: jetstream.AckAllPolicy,
	})
	require.NoError(t, err)

	wg := sync.WaitGroup{}

	h := TypedBatchHandlerFunc[testMessageType](func(ctx context.Context, batch jetmessage.TypedMessageBatch[testMessageType]) error {
		var last jetmessage.TypedMessage[testMessageType]
		for {
			msg, err := batch.Next(ctx)
			if errors.Is(err, io.EOF) {
				break
			}
			if errors.Is(err, context.Canceled) {
				break
			}
			require.NoError(t, err)
			last = msg
			wg.Done()
		}
		if last == nil {
			return nil
		}
		return last.Ack()
	})

	c := BatchConsumer{
		Fetcher:   consumer,
		Handler:   HandleBatchType[testMessageType](peanats.JsonCodec{}, h),
		BatchSize: 10,
	}
	err = c.Start()
	require.NoError(t, err)

	pub := peanats.NewTypedPublisher[testMessageType](peanats.JsonCodec{}, peanats.NewPublisher(nc)).WithSubject(t.Name())
	for i := 0; i < 10; i++ {
		wg.Add(1)
		err = pub.Publish(&testMessageType{Seq: uint(i)})
		require.NoError(t, err)
	}

	wg.Wait()
	c.Stop()
	require.NoError(t, c.Error())
}
