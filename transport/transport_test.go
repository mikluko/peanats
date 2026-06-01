package transport_test

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xtestutil"
	"github.com/mikluko/peanats/transport"
)

// TestSubscribeChan_UnsubscribeClosesChannel verifies the ownership model:
// peanats owns the send side of the caller's channel and closes it on
// Unsubscribe. The caller only ever receives. Run under -race to confirm no
// data race between teardown and the mirror goroutine's send.
func TestSubscribeChan_UnsubscribeClosesChannel(t *testing.T) {
	srv := xtestutil.Server(t)
	nc, err := nats.Connect(srv.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	conn := transport.New(nc)

	ch := make(chan peanats.Msg, 1)

	sub, err := conn.SubscribeChan(t.Context(), "test.subchan.close", ch)
	require.NoError(t, err)

	for range 10 {
		err = nc.Publish("test.subchan.close", []byte("test"))
		require.NoError(t, err)
	}
	err = nc.Flush()
	require.NoError(t, err)

	// Drain whatever has been delivered concurrently with teardown.
	go func() {
		for range ch { //nolint:revive // drain until closed
		}
	}()

	time.Sleep(50 * time.Millisecond)

	err = sub.Unsubscribe()
	require.NoError(t, err)

	// After Unsubscribe returns, the mirror goroutine has joined and ch is
	// closed; a receive must observe the closed channel, not block.
	select {
	case _, ok := <-ch:
		require.False(t, ok, "channel must be closed after Unsubscribe")
	case <-time.After(time.Second):
		t.Fatal("channel not closed after Unsubscribe")
	}
}

// TestSubscribeChan_StopReceivingThenUnsubscribe verifies teardown works even
// when the caller has stopped receiving (channel buffer full, mirror blocked
// on send). Unsubscribe must still join the mirror goroutine without leaking.
func TestSubscribeChan_StopReceivingThenUnsubscribe(t *testing.T) {
	srv := xtestutil.Server(t)
	nc, err := nats.Connect(srv.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	conn := transport.New(nc)

	// Unbuffered: the mirror goroutine blocks on send once one message is in
	// flight and the caller is not receiving.
	ch := make(chan peanats.Msg)

	sub, err := conn.SubscribeChan(t.Context(), "test.subchan.block", ch)
	require.NoError(t, err)

	for range 5 {
		err = nc.Publish("test.subchan.block", []byte("test"))
		require.NoError(t, err)
	}
	err = nc.Flush()
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	// Never received; Unsubscribe must still return promptly.
	done := make(chan error, 1)
	go func() { done <- sub.Unsubscribe() }()
	select {
	case err = <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Unsubscribe blocked: mirror goroutine leaked")
	}
}
