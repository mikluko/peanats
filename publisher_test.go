package peanats

import (
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func TestPublisher(t *testing.T) {
	m := msgPublisherMock{}
	m.On("PublishMsg", &nats.Msg{
		Subject: "the.parson.had.a.dog",
		Header: nats.Header{
			"the-parson": []string{"had a dog"},
		},
		Data: []byte("the parson had a dog"),
	}).Return(nil)

	p := publisher{
		PublisherMsg: &m,
		subject:      "the.parson.had.a.dog",
	}
	p.Header().Add("the-parson", "had a dog")
	err := p.Publish([]byte("the parson had a dog"))
	require.NoError(t, err)
}
