package peanats_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xmock/peanatsmock"
)

func TestArgumentMessageHandler(t *testing.T) {
	type testArg struct {
		Value string `json:"value"`
	}
	t.Run("happy path", func(t *testing.T) {
		h := peanats.MsgHandlerFromArgHandler(peanats.ArgHandlerFunc[testArg](func(ctx context.Context, arg peanats.Arg[testArg]) error {
			assert.Equal(t, "parson.had", arg.Subject())
			assert.Equal(t, peanats.Header{"x-parson": []string{"dog"}}, arg.Header())
			assert.Equal(t, &testArg{Value: "a dog"}, arg.Value())
			return nil
		}))
		m := peanatsmock.NewMsg(t)
		m.EXPECT().Subject().Return("parson.had")
		m.EXPECT().Header().Return(peanats.Header{"x-parson": []string{"dog"}})
		m.EXPECT().Data().Return([]byte(`{"value":"a dog"}`))
		err := h.HandleMsg(t.Context(), m)
		require.NoError(t, err)
	})
	t.Run("decode error", func(t *testing.T) {
		h := peanats.MsgHandlerFromArgHandler(peanats.ArgHandlerFunc[testArg](func(ctx context.Context, _ peanats.Arg[testArg]) error {
			panic("should not be called")
		}))
		m := peanatsmock.NewMsg(t)
		m.EXPECT().Header().Return(peanats.Header{"x-parson": []string{"dog"}})
		m.EXPECT().Data().Return([]byte(`{`))
		err := h.HandleMsg(t.Context(), m)
		require.Error(t, err)
		require.ErrorIs(t, err, peanats.ErrArgumentDecodeFailed)
	})
	t.Run("handler error", func(t *testing.T) {
		handlerErr := fmt.Errorf("test error")
		h := peanats.MsgHandlerFromArgHandler(peanats.ArgHandlerFunc[testArg](func(ctx context.Context, _ peanats.Arg[testArg]) error {
			return handlerErr
		}))
		m := peanatsmock.NewMsg(t)
		m.EXPECT().Header().Return(peanats.Header{"x-parson": []string{"dog"}})
		m.EXPECT().Data().Return([]byte(`{}`))
		err := h.HandleMsg(t.Context(), m)
		require.Error(t, err)
		assert.ErrorIs(t, err, handlerErr)
	})
}
