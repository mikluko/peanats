package testutil

import (
	"testing"
	"time"

	natsrv "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func NatsServer(tb testing.TB) *natsrv.Server {
	srv := must(natsrv.NewServer(&natsrv.Options{
		ServerName:    "test-server",
		Host:          "127.0.0.1",
		Port:          -1,
		WriteDeadline: time.Second * 10,
		JetStream:     true,
		StoreDir:      tb.TempDir(),
	}))
	go srv.Start()
	if !srv.ReadyForConnections(time.Second * 10) {
		tb.Fatal("nats server failed to start")
	}
	tb.Cleanup(srv.Shutdown)
	return srv
}

func NatsConn(tb testing.TB, srv *natsrv.Server, opts ...nats.Option) *nats.Conn {
	nc := must(nats.Connect(srv.ClientURL(), opts...))
	tb.Cleanup(nc.Close)
	return nc
}
