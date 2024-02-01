package peanats

import (
	"github.com/nats-io/nats.go"
)

// MakeDoneMiddleware makes a middleware that sends a message after handler call completion. That is useful to inform
// calling party that it shouldn't expect further messages in a case where handler sends multiple responses. Done won't
// be sent in case handler returns error. Consider wrapping handler with ErrorHandlerMiddleware deeper in the chain
// in case that isn't the desired behavior.
func MakeDoneMiddleware(msgpub PublisherMsg, opts ...DoneMiddlewareOption) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			err := next.Serve(pub, req)
			if err != nil {
				return err
			}
			msg := nats.NewMsg(req.Reply())
			for i := range opts {
				opts[i](msg)
			}
			if msg.Subject != "" {
				err = msgpub.PublishMsg(msg)
			}
			return err
		})
	}
}

type DoneMiddlewareOption func(msg *nats.Msg)

// DoneMiddlewareWithPayload option configures static payload to send with ack message
func DoneMiddlewareWithPayload(p []byte) DoneMiddlewareOption {
	return func(msg *nats.Msg) {
		msg.Data = p
	}
}

// DoneMiddlewareWithHeader option configures headers for the ack message
func DoneMiddlewareWithHeader(header nats.Header) DoneMiddlewareOption {
	return func(msg *nats.Msg) {
		for key := range header {
			for _, value := range header.Values(key) {
				msg.Header.Add(key, value)
			}
		}
	}
}
