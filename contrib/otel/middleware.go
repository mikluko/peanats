package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/mikluko/peanats"
)

type MiddlewareOption func(*middlewareParams)

type middlewareParams struct {
	tracer trace.Tracer
	attrs  []attribute.KeyValue
	header bool
	data   bool
}

func MiddlewareTracer(t trace.Tracer) MiddlewareOption {
	return func(p *middlewareParams) {
		p.tracer = t
	}
}

func MiddlewareAttributes(attrs []attribute.KeyValue) MiddlewareOption {
	return func(p *middlewareParams) {
		p.attrs = attrs
	}
}

func MiddlewareWithHeader(x bool) MiddlewareOption {
	return func(p *middlewareParams) {
		p.header = x
	}
}

func MiddlewareWithData(x bool) MiddlewareOption {
	return func(p *middlewareParams) {
		p.data = x
	}
}

func Middleware(opts ...MiddlewareOption) peanats.MsgMiddleware {
	p := middlewareParams{
		tracer: otel.Tracer("peanats/middleware"),
	}
	for _, opt := range opts {
		opt(&p)
	}
	return func(next peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
			attrs := append(p.attrs, MessageSubjectAttribute(msg.Subject()))
			if p.data {
				attrs = append(attrs, MessageDataAttribute(msg.Data()))
			}
			if p.header {
				attrs = append(attrs, MessageHeaderAttributes(msg.Header())...)
			}
			if msgx, ok := msg.(peanats.Metadatable); ok {
				meta, err := msgx.Metadata()
				if err != nil {
					attrs = append(attrs, MessageMetadataAttributes(meta)...)
				}
			}
			ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(msg.Header()))
			ctx, span := p.tracer.Start(ctx, "peanats/middleware",
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(attrs...),
			)
			defer span.End()
			return next.HandleMsg(ctx, msg)
		})
	}
}
