package peanats

import (
	"net/http"

	"github.com/nats-io/nats.go"
)

// MakeAckMiddleware makes a middleware that sends a message before passing control down to the handler. That might be
// beneficial for scenarios where client wants to immediately receive confirmation that request is being processed.
func MakeAckMiddleware(msgpub PublisherMsg, opts ...AckMiddlewareOption) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			msg := nats.NewMsg(req.Reply())
			for i := range opts {
				opts[i](msg)
			}
			if msg.Subject != "" {
				err := msgpub.PublishMsg(msg)
				if err != nil {
					return &Error{
						Code:    http.StatusInternalServerError,
						Message: "Failed to publish ack",
						Cause:   err,
					}
				}
			}
			return next.Serve(pub, req)
		})
	}
}

type AckMiddlewareOption func(msg *nats.Msg)

// AckMiddlewareWithPayload option configures static payload to send with ack message
func AckMiddlewareWithPayload(p []byte) AckMiddlewareOption {
	return func(msg *nats.Msg) {
		msg.Data = p
	}
}

// AckMiddlewareWithHeader option configures headers for the ack message
func AckMiddlewareWithHeader(header nats.Header) AckMiddlewareOption {
	return func(msg *nats.Msg) {
		for key := range header {
			for _, value := range header.Values(key) {
				msg.Header.Add(key, value)
			}
		}
	}
}
