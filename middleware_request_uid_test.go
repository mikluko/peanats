package peanats

import (
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	defer req.AssertExpectations(t)

	pub := new(publisherMock)
	defer pub.AssertExpectations(t)

	var f Handler
	f = HandlerFunc(func(pub Publisher, req Request) error {
		require.NoError(t, pub.(Publisher).Publish(nil))
		return nil
	})
	f = ChainMiddleware(f, RequestUIDMiddleware)

	reqHeader := make(nats.Header)
	req.On("Header").Return(reqHeader)

	pubHeader := make(nats.Header)
	pub.On("Header").Return(&pubHeader)
	pub.On("Publish", mock.Anything).Return(nil)

	err := f.Serve(pub, req)
	require.NoError(t, err)
	require.NotEqual(t, "", reqHeader.Get(HeaderRequestUID))
	require.NotEqual(t, "", pubHeader.Get(HeaderRequestUID))
}
