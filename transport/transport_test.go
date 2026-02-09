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

func TestMirror_ClosedChannelRecovery(t *testing.T) {
	srv := xtestutil.Server(t)
	nc, err := nats.Connect(srv.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	conn := transport.New(nc)

	ch := make(chan peanats.Msg, 1)

	sub, err := conn.SubscribeChan(t.Context(), "test.mirror.panic", ch)
	require.NoError(t, err)

	for range 10 {
		err = nc.Publish("test.mirror.panic", []byte("test"))
		require.NoError(t, err)
	}
	err = nc.Flush()
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	close(ch)

	err = sub.Unsubscribe()
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
}
