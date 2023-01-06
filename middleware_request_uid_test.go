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
	req := new(requestMock)
	//defer req.AssertExpectations(t)

	pub := new(publisherMock)
	//defer pub.AssertExpectations(t)

	gen := new(uidGeneratorMock)
	//defer gen.AssertExpectations(t)

	uid := nuid.Next()
	gen.On("Next").Return(uid)

	var f Handler
	f = HandlerFunc(func(pub Publisher, req Request) error {
		require.NoError(t, pub.(Acker).Ack(nil))
		require.NoError(t, pub.(Publisher).Publish(nil))
		return nil
	})
	f = ChainMiddleware(f, MakeRequestUIDMiddleware(gen))

	reqHeader := make(nats.Header)
	req.On("Header").Return(&reqHeader)

	pubHeader := make(nats.Header)
	pub.On("Header").Return(&pubHeader)
	pub.On("Ack", mock.Anything).Return(nil)
	pub.On("Publish", mock.Anything).Return(nil)

	err := f.Serve(pub, req)
	require.NoError(t, err)
	require.Equal(t, uid, reqHeader.Get(HeaderRequestUID))
	require.Equal(t, uid, pubHeader.Get(HeaderRequestUID))
}
