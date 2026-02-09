package xtestutil

import (
	"testing"
	"time"

	natsrv "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats/transport"
)

func Server(tb testing.TB) *natsrv.Server {
	srv := Must(natsrv.NewServer(&natsrv.Options{
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

func Conn(tb testing.TB, srv *natsrv.Server, opts ...nats.Option) transport.Conn {
	conn := Must(transport.Wrap(Must(nats.Connect(srv.ClientURL(), opts...))))
	tb.Cleanup(conn.Close)
	return conn
}
