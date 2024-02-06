package peanats

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublishSubjectMiddleware(t *testing.T) {
	f := ChainMiddleware(
		HandlerFunc(func(pub Publisher, req Request) error {
			return pub.Publish([]byte("the parson had a dog"))
		}),
		MakePublishSubjectMiddleware("the.parson.had.a.dog"),
	)

	pub := new(publisherMock)
	pub.On("WithSubject", "the.parson.had.a.dog").Return(pub).Once()
	pub.On("Publish", []byte("the parson had a dog")).Return(nil).Once()

	req := new(requestMock)

	err := f.Serve(pub, req)
	require.NoError(t, err)

	req.AssertExpectations(t)
	pub.AssertExpectations(t)
}
