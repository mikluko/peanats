package peanats

import (
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func TestPublishSubjectMiddleware(t *testing.T) {
	f := ChainMiddleware(
		HandlerFunc(func(pub Publisher, req Request) error {
			return pub.Publish([]byte("the parson had a dog"))
		}),
		MakePublishSubjectMiddleware("the.parson.had.a.dog"),
	)

	req := new(requestMock)
	defer req.AssertExpectations(t)

	pub := new(publisherMock)
	defer pub.AssertExpectations(t)

	pub.On("Header").Return(&nats.Header{
		"the-parson": []string{"had a dog"},
	}).Once()

	pub.On("PublishMsg", &nats.Msg{
		Subject: "the.parson.had.a.dog",
		Header: nats.Header{
			"the-parson": []string{"had a dog"},
		},
		Data: []byte("the parson had a dog"),
	}).Return(nil).Once()

	err := f.Serve(pub, req)
	require.NoError(t, err)

}
