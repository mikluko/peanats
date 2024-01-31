package peanats

import (
	"context"
	"fmt"
	"testing"

	"github.com/nats-io/nats.go"
)

func BenchmarkSimple(b *testing.B) {
	ns := RunNats(b)

	srvConn, err := nats.Connect(ns.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer srvConn.Close()

	srv := Server{
		Conn:           NATS(srvConn),
		ListenSubjects: []string{"test.requests"},
		Handler: HandlerFunc(func(pub Publisher, req Request) (err error) {
			return pub.Publish([]byte("test"))
		}),
	}
	err = srv.Start()
	if err != nil {
		b.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cliConn, err := nats.Connect(ns.Addr().String())

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cliConn.RequestWithContext(ctx, "test.requests", []byte("test"))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkTypedJson(b *testing.B) {
	ns := RunNats(b)

	srvConn, err := nats.Connect(ns.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(srvConn.Close)

	type argument struct {
		Arg string `json:"arg"`
	}
	type result struct {
		Res string `json:"res"`
	}

	p := TypedHandlerFunc[argument, result](
		func(pub TypedPublisher[result], req TypedRequest[argument]) error {
			res := result{Res: req.Payload().Arg}
			return pub.Publish(&res)
		},
	)

	srv := Server{
		Conn:           NATS(srvConn),
		ListenSubjects: []string{"test.requests"},
		Handler:        Typed[argument, result](&JsonCodec{}, p),
	}
	err = srv.Start()
	if err != nil {
		b.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cliConn, err := nats.Connect(ns.Addr().String())
	b.Cleanup(cliConn.Close)

	data, _ := (&JsonCodec{}).Encode(&argument{Arg: "test"})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cliConn.RequestWithContext(ctx, "test.requests", data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkServeMux(b *testing.B) {
	var benchmark = func(b *testing.B, num int) {
		mux := ServeMux{}
		_ = mux.HandleFunc(func(pub Publisher, req Request) error {
			return pub.Publish([]byte("hello"))
		}, "foo")
		for i := 0; i < num; i++ {
			_ = mux.HandleFunc(func(pub Publisher, req Request) error {
				b.Fatal("should not be called")
				return nil
			}, fmt.Sprintf("bar-%07d", i))
		}
		req := new(requestMock)
		req.On("Subject").Return("foo").Times(b.N)

		pub := new(publisherMock)
		pub.On("Publish", []byte("hello")).Return(nil).Times(b.N)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = mux.Serve(pub, req)
			}
		})
	}
	b.Run("1", func(b *testing.B) { benchmark(b, 1) })
	b.Run("10", func(b *testing.B) { benchmark(b, 10) })
	b.Run("100", func(b *testing.B) { benchmark(b, 100) })
}
