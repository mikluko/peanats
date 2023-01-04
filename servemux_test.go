package peanats

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
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

		pub := new(publisherAckerMock)
		defer pub.AssertExpectations(t)

		req.On("Subject").Return("foo")
		pub.On("Publish", []byte("hello")).Return(nil)

		err := mux.Serve(pub, req)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		req := new(requestMock)
		defer req.AssertExpectations(t)

		pub := new(publisherAckerMock)
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

func BenchmarkServeMux(b *testing.B) {
	var benchmark = func(b *testing.B, num int) {
		mux := ServeMux{}
		_ = mux.HandleFunc(func(pub Publisher, req Request) error {
			return pub.Publish([]byte("hello"))
		}, "foo")
		for i := 0; i < num; i++ {
			_ = mux.HandleFunc(func(pub Publisher, req Request) error {
				b.Fatal("should not be called")
				return nil
			}, fmt.Sprintf("bar-%07d", i))
		}
		req := new(requestMock)
		req.On("Subject").Return("foo").Times(b.N)

		pub := new(publisherAckerMock)
		pub.On("Publish", []byte("hello")).Return(nil).Times(b.N)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = mux.Serve(pub, req)
		}
	}
	b.Run("1", func(b *testing.B) { benchmark(b, 1) })
	b.Run("10", func(b *testing.B) { benchmark(b, 10) })
	b.Run("100", func(b *testing.B) { benchmark(b, 100) })
	b.Run("1000", func(b *testing.B) { benchmark(b, 1000) })
	b.Run("10000", func(b *testing.B) { benchmark(b, 10000) })
}
