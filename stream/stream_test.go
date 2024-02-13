package stream

import (
	"context"
	"testing"
	"time"

	natsrv "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
)

func benchmarkStartNatsServer(b *testing.B) *natsrv.Server {
	srv, err := natsrv.NewServer(&natsrv.Options{
		ServerName:    "test-server",
		Host:          "127.0.0.1",
		Port:          -1,
		WriteDeadline: time.Second * 1,
	})
	if err != nil {
		b.Fatal(err)
	}

	go srv.Start()
	if !srv.ReadyForConnections(time.Second * 10) {
		b.Fatal("nats server failed to start")
	}
	b.Cleanup(srv.Shutdown)

	return srv
}

func benchmarkStream(b *testing.B) {
	srv := benchmarkStartNatsServer(b)

	pc, err := nats.Connect(srv.ClientURL())
	if err != nil {
		b.Fatal(err)
	}
	ps := peanats.Server{
		Conn: peanats.NATS(pc),
		Handler: peanats.ChainMiddleware(
			peanats.HandlerFunc(func(pub peanats.Publisher, rq peanats.Request) error {
				return pub.Publish(rq.Data())
			}),
			Middleware,
		),
		ListenSubjects: []string{"test.stream"},
	}
	err = ps.Start()
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(ps.Shutdown)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b.ReportAllocs()
	b.ResetTimer()

	c := NewClient(peanats.NATS(pc))

	for i := 0; i < b.N; i++ {
		rcv, err := c.Start(ctx, "test.stream", []byte("the parson had a dog"))
		if err != nil {
			b.Fatal(err)
		}
		_, err = rcv.ReceiveAll(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkTypedStream(b *testing.B) {
	srv := benchmarkStartNatsServer(b)

	type Arg struct {
		Parson string `json:"parson"`
	}

	type Res struct {
		Dog string `json:"dog"`
	}

	pc, err := nats.Connect(srv.ClientURL())
	if err != nil {
		b.Fatal(err)
	}
	ps := peanats.Server{
		Conn: peanats.NATS(pc),
		Handler: peanats.ChainMiddleware(
			peanats.Typed[Arg, Res](
				peanats.JsonCodec{},
				peanats.TypedHandlerFunc[Arg, Res](
					func(pub peanats.TypedPublisher[Res], rq peanats.TypedRequest[Arg]) error {
						return pub.Publish(&Res{Dog: "was owned by the parson"})
					},
				),
			),
			Middleware,
		),
		ListenSubjects: []string{"test.stream"},
	}
	err = ps.Start()
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(ps.Shutdown)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b.ReportAllocs()
	b.ResetTimer()

	c := NewTypedClient[Arg, Res](peanats.NATS(pc))

	for i := 0; i < b.N; i++ {
		rcv, err := c.Start(ctx, "test.stream", &Arg{Parson: "had a dog"})
		if err != nil {
			b.Fatal(err)
		}
		_, err = rcv.ReceiveAll(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStream(b *testing.B) {
	b.Run("raw", benchmarkStream)
	b.Run("typed", benchmarkTypedStream)
}
