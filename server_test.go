package peanats

import (
	"context"
	"sync"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats/testutil"
)

func TestServer(t *testing.T) {
	ns := testutil.NatsServer(t)

	nc, err := nats.Connect(ns.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Close()

	h := HandlerFunc(func(pub Publisher, req Request) error {
		require.Equal(t, []byte("test"), req.Data())
		require.Equal(t, nats.Header{}, req.Header()) // no headers should not return nil
		return nil
	})

	wg := sync.WaitGroup{}
	mw := Middleware(func(next Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			defer wg.Done()
			return next.Serve(pub, req)
		})
	})

	srv := Server{
		Conn:           NATS(nc),
		ListenSubjects: []string{"test.requests"},
		Handler:        ChainMiddleware(h, mw),
		Concurrency:    10,
	}
	err = srv.Start()
	require.NoError(t, err)

	for i := 0; i < srv.Concurrency; i++ {
		wg.Add(1)
		err = nc.Publish("test.requests", []byte("test"))
		require.NoError(t, err)
	}
	wg.Wait()

	srv.Shutdown()
	srv.Wait()
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
