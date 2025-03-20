package misc

import (
	"context"

	"github.com/mikluko/peanats"
)

// DefaultContentTypeMiddleware will inject specified content type into incoming
// messages that don't specify one.
func DefaultContentTypeMiddleware(c peanats.ContentType) peanats.MsgMiddleware {
	return func(next peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
			if msg.Header().Get(peanats.HeaderContentType) == "" {
				msg.Header().Set(peanats.HeaderContentType, c.String())
			}
			return next.HandleMsg(ctx, msg)
		})
	}
}
