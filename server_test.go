package peanats

import (
	"context"
	natsrv "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func RunNats(tb testing.TB) *natsrv.Server {
	srv := must(natsrv.NewServer(&natsrv.Options{
		ServerName:    "test-server",
		Host:          "127.0.0.1",
		Port:          -1,
		WriteDeadline: time.Second * 10,
	}))

	go srv.Start()
	if !srv.ReadyForConnections(time.Second * 10) {
		tb.Fatal("nats server failed to start")
	}
	tb.Cleanup(srv.Shutdown)
	return srv
}

func TestServer(t *testing.T) {
	ns := RunNats(t)

	nc, err := nats.Connect(ns.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := Server{
		Conn:           nc,
		ListenSubjects: []string{"test.requests"},
		Handler: ChainMiddleware(HandlerFunc(func(pub Publisher, req Request) error {
			require.Equal(t, []byte("test"), req.Data())
			ack := pub.(Acker)
			err := ack.Ack(nil)
			require.NoError(t, err)
			return pub.Publish([]byte("test"))
		}), MakePublishSubjectMiddleware(nc, "test.results")),
	}
	err = srv.Start()
	require.NoError(t, err)

	sub, err := nc.SubscribeSync("test.results")
	require.NoError(t, err)
	_ = sub.AutoUnsubscribe(1)

	_, err = nc.RequestWithContext(ctx, "test.requests", []byte("test"))
	require.NoError(t, err)

	msg, err := sub.NextMsgWithContext(ctx)
	require.NoError(t, err)
	require.Equal(t, []byte("test"), msg.Data)

	err = srv.Shutdown()
	require.NoError(t, err)

	err = srv.Wait()
	require.NoError(t, err)
}

type publisherMock struct {
	mock.Mock
}

func (p *publisherMock) Header() *nats.Header {
	args := p.Called()
	return args.Get(0).(*nats.Header)
}

func (p *publisherMock) Publish(data []byte) error {
	args := p.Called(data)
	return args.Error(0)
}

type publisherAckerMock struct {
	publisherMock
}

func (p *publisherAckerMock) Ack(data []byte) error {
	args := p.Called(data)
	return args.Error(0)
}

type requestMock struct {
	mock.Mock
}

func (r *requestMock) Context() context.Context {
	args := r.Mock.Called()
	return args.Get(0).(context.Context)
}

func (r *requestMock) WithContext(ctx context.Context) Request {
	args := r.Mock.Called(ctx)
	return args.Get(0).(Request)
}

func (r *requestMock) Subject() string {
	args := r.Mock.Called()
	return args.Get(0).(string)
}

func (r *requestMock) Header() *nats.Header {
	args := r.Mock.Called()
	return args.Get(0).(*nats.Header)
}

func (r *requestMock) Data() []byte {
	args := r.Mock.Called()
	return args.Get(0).([]byte)
}
