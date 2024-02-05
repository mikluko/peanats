package peanats

import (
	"context"
	"fmt"
	"github.com/mikluko/peanats/examples/protojson/api"
	"testing"

	"github.com/nats-io/nats.go"
)

func BenchmarkServer(b *testing.B) {
	ns := RunNats(b)

	sc, err := nats.Connect(ns.Addr().String())
	if err != nil {
		b.Fatal(err)
	}
	defer sc.Close()

	srv := Server{
		Conn:           NATS(sc),
		ListenSubjects: []string{"test.simple"},
		Handler: ChainMiddleware(
			HandlerFunc(func(pub Publisher, req Request) (err error) {
				err = pub.Publish(req.Data())
				if err != nil {
					return err
				}
				return nil
			}),
		),
	}
	err = srv.Start()
	if err != nil {
		b.Fatal(err)
	}

	cc, err := nats.Connect(ns.Addr().String())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg, err := cc.RequestWithContext(ctx, "test.simple", []byte("the parson had a dog"))
		if err != nil {
			b.Fatal(err)
		}
		if string(msg.Data) != "the parson had a dog" {
			b.Fatalf("unexpected response: %s", msg.Data)
		}
	}
}

func BenchmarkRaw(b *testing.B) {
	type argument struct {
		Arg string `json:"arg"`
	}
	type result struct {
		Res string `json:"res"`
	}

	handler := Typed[argument, result](&JsonCodec{}, TypedHandlerFunc[argument, result](
		func(pub TypedPublisher[result], req TypedRequest[argument]) error {
			res := result{Res: req.Payload().Arg}
			return pub.Publish(&res)
		},
	))

	req := new(requestMock)
	req.On("Subject").Return("foo").Times(b.N)
	req.On("Data").Return([]byte("{\"arg\":\"the parson had a dog\"}")).Times(b.N)

	pub := new(publisherMock)
	pub.On("Publish", []byte("{\"res\":\"the parson had a dog\"}")).Return(nil).Times(b.N)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := handler.Serve(pub, req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func benchmarkTyped[RQ, RS any](b *testing.B, codec Codec, rq *RQ, rs *RS) {
	handler := Typed[RQ, RS](codec, TypedHandlerFunc[RQ, RS](
		func(pub TypedPublisher[RS], req TypedRequest[RQ]) error {
			return pub.Publish(rs)
		},
	))

	rqp, err := codec.Encode(rq)
	if err != nil {
		b.Fatalf("request encode failed: %s", err)
	}
	req := new(requestMock)
	req.On("Subject").Return("foo").Times(b.N)
	req.On("Data").Return(rqp).Times(b.N)

	rsp, _ := codec.Encode(rs)
	if err != nil {
		b.Fatalf("response encode failed: %s", err)
	}
	pub := new(publisherMock)
	pub.On("Publish", rsp).Return(nil).Times(b.N)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := handler.Serve(pub, req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkTyped(b *testing.B) {
	b.Run("json", func(b *testing.B) {
		benchmarkTyped(b, &JsonCodec{}, &api.Argument{Arg: "the parson had a dog"}, &api.Result{Res: "the parson had a dog"})
	})
	b.Run("proto", func(b *testing.B) {
		benchmarkTyped(b, &ProtoCodec{}, &api.Argument{Arg: "the parson had a dog"}, &api.Result{Res: "the parson had a dog"})
	})
	b.Run("protojson", func(b *testing.B) {
		benchmarkTyped(b, &ProtojsonCodec{}, &api.Argument{Arg: "the parson had a dog"}, &api.Result{Res: "the parson had a dog"})
	})
	b.Run("prototext", func(b *testing.B) {
		benchmarkTyped(b, &PrototextCodec{}, &api.Argument{Arg: "the parson had a dog"}, &api.Result{Res: "the parson had a dog"})
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
		b.ReportAllocs()
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
