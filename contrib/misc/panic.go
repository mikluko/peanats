package misc

import (
	"context"

	"github.com/mikluko/peanats"
)

// PanicCallbackMiddleware is a middleware that intercepts panics during message handling
func PanicCallbackMiddleware(cb func(any)) peanats.MsgMiddleware {
	return func(h peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, m peanats.Msg) error {
			defer func() {
				if r := recover(); r != nil {
					cb(r)
				}
			}()
			return h.HandleMsg(ctx, m)
		})
	}
}
