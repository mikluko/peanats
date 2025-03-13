package peanats_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
)

func TestClient(t *testing.T) {
	type request struct {
		Foo string `json:"foo"`
	}
	type response struct {
		Bar string `json:"bar"`
	}
	t.Run("happy path", func(t *testing.T) {
		nc := newNatsRequesterMock(t)
		nc.EXPECT().
			RequestMsgWithContext(mock.Anything, mock.Anything).
			Run(func(_ context.Context, msg *nats.Msg) {
				assert.Equal(t, "parson.had", msg.Subject)
				assert.JSONEq(t, `{"foo": "a dog"}`, string(msg.Data))
			}).
			Return(&nats.Msg{Data: []byte(`{"bar": "a dog"}`)}, nil).Once()
		c := peanats.NewClient[request, response](nc)
		rs, err := c.Request(context.Background(), "parson.had", &request{Foo: "a dog"})
		require.NoError(t, err)
		require.NotNil(t, rs)
		assert.Equal(t, &response{Bar: "a dog"}, rs.Payload())
	})
	t.Run("decode error", func(t *testing.T) {
		nc := newNatsRequesterMock(t)
		nc.EXPECT().
			RequestMsgWithContext(mock.Anything, mock.Anything).
			Return(&nats.Msg{Data: []byte(`{`)}, nil).Once()
		c := peanats.NewClient[request, response](nc)
		rs, err := c.Request(context.Background(), "parson.had", &request{Foo: "a dog"})
		require.Error(t, err)
		require.Nil(t, rs)
	})
	t.Run("some error", func(t *testing.T) {
		testErr := errors.New("parson had a dog")
		nc := newNatsRequesterMock(t)
		nc.EXPECT().
			RequestMsgWithContext(mock.Anything, mock.Anything).
			Return(nil, testErr).Once()
		c := peanats.NewClient[request, response](nc)
		rs, err := c.Request(context.Background(), "parson.had", &request{Foo: "a dog"})
		require.Error(t, err)
		assert.ErrorIs(t, err, testErr)
		assert.Nil(t, rs)
	})
}
