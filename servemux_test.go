package peanats

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServeMux(t *testing.T) {

	mux := ServeMux{}
	_ = mux.HandleFunc(func(pub Publisher, req Request) error {
		return pub.Publish([]byte("hello"))
	}, "foo")
	_ = mux.HandleFunc(func(pub Publisher, req Request) error {
		t.Fatal("should not be called")
		return nil
	}, "bar")

	t.Run("ok", func(t *testing.T) {
		req := new(requestMock)
		defer req.AssertExpectations(t)

		pub := new(publisherMock)
		defer pub.AssertExpectations(t)

		req.On("Subject").Return("foo")
		pub.On("Publish", []byte("hello")).Return(nil)

		err := mux.Serve(pub, req)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		req := new(requestMock)
		defer req.AssertExpectations(t)

		pub := new(publisherMock)
		defer pub.AssertExpectations(t)

		req.On("Subject").Return("baz")

		err := mux.Serve(pub, req)
		require.Error(t, err)

		owl := new(Error)
		require.True(t, errors.As(err, &owl))
		require.Equal(t, http.StatusNotFound, owl.Code)
		require.Equal(t, http.StatusText(http.StatusNotFound), owl.Message)
	})
}
