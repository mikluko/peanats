package peaclient_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats/internal/xmock/xmsgmock"
	"github.com/mikluko/peanats/internal/xmock/xnatsmock"
	"github.com/mikluko/peanats/peaclient"
	"github.com/mikluko/peanats/xmsg"
)

func TestClient(t *testing.T) {
	type request struct {
		Foo string `json:"foo"`
	}
	type response struct {
		Bar string `json:"bar"`
	}
	t.Run("happy path", func(t *testing.T) {
		msg := xmsgmock.NewMsg(t)
		msg.EXPECT().Data().Return([]byte(`{"bar": "a dog"}`)).Once()
		msg.EXPECT().Header().Return(xmsg.Header{})
		nc := xnatsmock.NewConnection(t)
		nc.EXPECT().
			Request(mock.Anything, mock.Anything).
			Run(func(_ context.Context, msg xmsg.Msg) {
				assert.Equal(t, "parson.had", msg.Subject())
				assert.JSONEq(t, `{"foo": "a dog"}`, string(msg.Data()))
			}).
			Return(msg, nil).Once()
		c := peaclient.New[request, response](nc)
		rs, err := c.Request(context.Background(), "parson.had", &request{Foo: "a dog"})
		require.NoError(t, err)
		require.NotNil(t, rs)
		assert.Equal(t, &response{Bar: "a dog"}, rs.Value())
	})
	t.Run("decode error", func(t *testing.T) {
		msg := xmsgmock.NewMsg(t)
		msg.EXPECT().Data().Return([]byte(`{`)).Once()
		msg.EXPECT().Header().Return(xmsg.Header{})
		nc := xnatsmock.NewConnection(t)
		nc.EXPECT().
			Request(mock.Anything, mock.Anything).
			Return(msg, nil).Once()
		c := peaclient.New[request, response](nc)
		rs, err := c.Request(context.Background(), "parson.had", &request{Foo: "a dog"})
		require.Error(t, err)
		require.Nil(t, rs)
	})
	t.Run("some error", func(t *testing.T) {
		testErr := errors.New("parson had a dog")
		nc := xnatsmock.NewConnection(t)
		nc.EXPECT().
			Request(mock.Anything, mock.Anything).
			Return(nil, testErr).Once()
		c := peaclient.New[request, response](nc)
		rs, err := c.Request(context.Background(), "parson.had", &request{Foo: "a dog"})
		require.Error(t, err)
		assert.ErrorIs(t, err, testErr)
		assert.Nil(t, rs)
	})
}
