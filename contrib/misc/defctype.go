package misc

import (
	"context"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/codec"
)

// DefaultContentTypeMiddleware will inject specified content type into incoming
// messages that don't specify one.
func DefaultContentTypeMiddleware(c codec.ContentType) peanats.MsgMiddleware {
	return func(next peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
			if msg.Header().Get(codec.HeaderContentType) == "" {
				msg.Header().Set(codec.HeaderContentType, c.String())
			}
			return next.HandleMsg(ctx, msg)
		})
	}
}
