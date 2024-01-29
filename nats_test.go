package peanats

import (
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
)

type msgPublisherMock struct {
	mock.Mock
}

func (m *msgPublisherMock) PublishMsg(msg *nats.Msg) error {
	args := m.Called(msg)
	return args.Error(0)
}
