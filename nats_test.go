package peanats_test

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xtestutil"
)

func TestMirror_ClosedChannelRecovery(t *testing.T) {
	// This test verifies that the mirror goroutine doesn't panic when the
	// consumer closes the channel while messages are still being received.
	// See: https://github.com/mikluko/peanats/issues/19

	srv := xtestutil.Server(t)
	nc, err := nats.Connect(srv.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	conn := peanats.NewConnection(nc)

	// Create a small buffered channel
	ch := make(chan peanats.Msg, 1)

	sub, err := conn.SubscribeChan(t.Context(), "test.mirror.panic", ch)
	require.NoError(t, err)

	// Publish several messages to fill the channel and NATS internal buffer
	for range 10 {
		err = nc.Publish("test.mirror.panic", []byte("test"))
		require.NoError(t, err)
	}
	err = nc.Flush()
	require.NoError(t, err)

	// Give time for messages to arrive at the subscription
	time.Sleep(50 * time.Millisecond)

	// Close the channel while messages are still pending - this would panic before the fix
	close(ch)

	// Unsubscribe to stop delivery
	err = sub.Unsubscribe()
	require.NoError(t, err)

	// Give time for the mirror goroutine to process remaining messages and exit
	time.Sleep(50 * time.Millisecond)

	// If we reach here without panic, the test passes
}
