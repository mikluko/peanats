package peanats

import (
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type publisherMock struct {
	mock.Mock
}

func (p *publisherMock) Subject() string {
	args := p.Called()
	return args.String(0)
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

func TestPublisher(t *testing.T) {
	m := publisherMock{}
	m.On("PublishMsg", &nats.Msg{
		Subject: "the.parson.had.a.dog",
		Header: nats.Header{
			"the-parson": []string{"had a dog"},
		},
		Data: []byte("the parson had a dog"),
	}).Return(nil)

	p := publisherImpl{
		PublisherMsg: &m,
		subject:      "the.parson.had.a.dog",
	}
	p.Header().Add("the-parson", "had a dog")
	err := p.Publish([]byte("the parson had a dog"))
	require.NoError(t, err)
}
