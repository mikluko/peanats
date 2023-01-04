package peanats

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type uidGeneratorMock struct {
	mock.Mock
}

func (t *uidGeneratorMock) Next() string {
	args := t.Called()
	return args.String(0)
}

func TestEnsureUIDMiddleware(t *testing.T) {
	rq := new(requestMock)
	defer rq.AssertExpectations(t)

	pub := new(publisherMock)
	defer pub.AssertExpectations(t)

	gen := new(uidGeneratorMock)
	defer gen.AssertExpectations(t)

	var f Handler
	f = HandlerFunc(func(pub Publisher, req Request) error { return nil })
	f = ChainMiddleware(f, MakeRequestUIDMiddleware(gen))

	header := make(nats.Header)
	rq.On("Header").Return(&header)

	uid := nuid.Next()
	gen.On("Next").Return(uid)

	err := f.Serve(pub, rq)
	require.NoError(t, err)
	require.Equal(t, uid, header.Get(HeaderRequestUID))
}
