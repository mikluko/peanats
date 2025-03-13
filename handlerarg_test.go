package peanats_test

import (
	"context"
	"fmt"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
)

func TestArgumentMessageHandler(t *testing.T) {
	type testArg struct {
		Value string `json:"value"`
	}
	t.Run("happy path", func(t *testing.T) {
		h := peanats.ArgumentMessageHandler(peanats.ArgumentHandlerFunc[testArg](func(ctx context.Context, arg peanats.Argument[testArg]) error {
			assert.Equal(t, "parson.had", arg.Subject())
			assert.Equal(t, textproto.MIMEHeader{"x-parson": []string{"dog"}}, arg.Header())
			assert.Equal(t, &testArg{Value: "a dog"}, arg.Value())
			return nil
		}))
		m := newMessageMock(t)
		m.EXPECT().Subject().Return("parson.had")
		m.EXPECT().Header().Return(textproto.MIMEHeader{"x-parson": []string{"dog"}})
		m.EXPECT().Data().Return([]byte(`{"value":"a dog"}`))
		err := h.Handle(t.Context(), m)
		require.NoError(t, err)
	})
	t.Run("decode error", func(t *testing.T) {
		h := peanats.ArgumentMessageHandler(peanats.ArgumentHandlerFunc[testArg](func(ctx context.Context, _ peanats.Argument[testArg]) error {
			panic("should not be called")
		}))
		m := newMessageMock(t)
		m.EXPECT().Header().Return(textproto.MIMEHeader{"x-parson": []string{"dog"}})
		m.EXPECT().Data().Return([]byte(`{`))
		err := h.Handle(t.Context(), m)
		require.Error(t, err)
		require.ErrorIs(t, err, peanats.ErrArgumentDecodeFailed)
	})
	t.Run("handler error", func(t *testing.T) {
		handlerErr := fmt.Errorf("test error")
		h := peanats.ArgumentMessageHandler(peanats.ArgumentHandlerFunc[testArg](func(ctx context.Context, _ peanats.Argument[testArg]) error {
			return handlerErr
		}))
		m := newMessageMock(t)
		m.EXPECT().Header().Return(textproto.MIMEHeader{"x-parson": []string{"dog"}})
		m.EXPECT().Data().Return([]byte(`{}`))
		err := h.Handle(t.Context(), m)
		require.Error(t, err)
		assert.ErrorIs(t, err, handlerErr)
	})
}
