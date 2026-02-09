package logging

import (
	"context"
	"time"

	"github.com/mikluko/peanats"
)

// AccessLogMiddleware is a middleware that logs the message subject and headers
func AccessLogMiddleware(log Logger) peanats.MsgMiddleware {
	return func(h peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, m peanats.Msg) error {
			t := time.Now()
			err := h.HandleMsg(ctx, m)
			if err != nil {
				return err
			}
			log.Log(ctx, "done", "subject", m.Subject(), "header", m.Header(), "latency", time.Since(t))
			return nil
		})
	}
}

// ErrorLogMiddleware is a middleware that logs errors encountered while handling messages
func ErrorLogMiddleware(log Logger) peanats.MsgMiddleware {
	return func(h peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, m peanats.Msg) error {
			err := h.HandleMsg(ctx, m)
			if err != nil {
				log.Log(ctx, "error", "message", err.Error())
			}
			return nil
		})
	}
}
