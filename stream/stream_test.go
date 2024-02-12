package stream

import (
	"context"
	"testing"
	"time"

	natsrv "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"

	"github.com/mikluko/peanats"
)

func BenchmarkStream(b *testing.B) {
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

	for i := 0; i < b.N; i++ {
		c := NewClient(peanats.NATS(pc))
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

type publisherMock struct {
	mock.Mock
}

func (p *publisherMock) Subject() string {
	args := p.Called()
	return args.String(0)
}

func (p *publisherMock) WithSubject(subj string) peanats.Publisher {
	args := p.Called(subj)
	return args.Get(0).(peanats.Publisher)
}

func (p *publisherMock) Header() *nats.Header {
	args := p.Called()
	return args.Get(0).(*nats.Header)
}

func (p *publisherMock) Publish(data []byte) error {
	args := p.Called(data)
	return args.Error(0)
}

func (p *publisherMock) PublishMsg(msg *nats.Msg) error {
	args := p.Called(msg)
	return args.Error(0)
}

type requestMock struct {
	mock.Mock
}

func (r *requestMock) Context() context.Context {
	args := r.Mock.Called()
	return args.Get(0).(context.Context)
}

func (r *requestMock) Subject() string {
	args := r.Mock.Called()
	return args.Get(0).(string)
}

func (r *requestMock) Reply() string {
	args := r.Mock.Called()
	return args.Get(0).(string)
}

func (r *requestMock) Header() nats.Header {
	args := r.Mock.Called()
	return args.Get(0).(nats.Header)
}

func (r *requestMock) Data() []byte {
	args := r.Mock.Called()
	return args.Get(0).([]byte)
}
