package misc

import (
	"context"

	"github.com/mikluko/peanats"
)

// SwallowErrMiddleware is a middleware that swallows errors from message handlers
// and always returns nil, effectively ignoring any errors that occur during message processing.
// The intention is to allow message processing to continue without interruption,
// even if the handler encounters an error.
//
// CAUTION: This middleware should probably be the outermost at all times and only used when you
// are sure that errors are properly handled deeper in the chain.
func SwallowErrMiddleware() peanats.MsgMiddleware {
	return func(h peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, m peanats.Msg) error {
			// Call the handler and ignore any errors
			_ = h.HandleMsg(ctx, m)
			return nil // Always return nil to swallow the error
		})
	}
}
