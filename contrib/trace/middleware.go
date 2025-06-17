package trace

import (
	"context"
	"fmt"
	"net/textproto"

	"github.com/mikluko/peanats"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// MiddlewareOption configures trace middleware
type MiddlewareOption func(*middlewareOptions)

type middlewareOptions struct {
	tracer         trace.Tracer
	spanKind       trace.SpanKind
	spanName       string
	spanAttributes []attribute.KeyValue
	withHeaders    bool
	withData       bool
}

// MiddlewareWithTracer sets the tracer for the middleware
func MiddlewareWithTracer(tracer trace.Tracer) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.tracer = tracer
	}
}

// MiddlewareWithSpanKind sets the span kind for traces
func MiddlewareWithSpanKind(kind trace.SpanKind) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.spanKind = kind
	}
}

// MiddlewareWithSpanName sets the span name for traces
func MiddlewareWithSpanName(name string) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.spanName = name
	}
}

// MiddlewareWithSpanAttributes adds attributes to the span created by the middleware
func MiddlewareWithSpanAttributes(attrs ...attribute.KeyValue) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.spanAttributes = append(o.spanAttributes, attrs...)
	}
}

// MiddlewareWithAttributes adds attributes to all spans created by the middleware
func MiddlewareWithAttributes(attrs ...attribute.KeyValue) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.spanAttributes = append(o.spanAttributes, attrs...)
	}
}

// MiddlewareWithHeaders enables trace context extraction from message headers
func MiddlewareWithHeaders(v bool) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.withHeaders = v
	}
}

// MiddlewareWithData enables recording message data as span attributes
func MiddlewareWithData(v bool) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.withData = v
	}
}

// Middleware creates a NATS message middleware that handles OpenTelemetry trace propagation
func Middleware(opts ...MiddlewareOption) peanats.MsgMiddleware {
	cfg := &middlewareOptions{
		tracer:   otel.Tracer("peanats"),
		spanName: "peanats.handle",
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return func(next peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
			attrs := append(cfg.spanAttributes,
				attribute.String("nats.subject", msg.Subject()),
			)
			if cfg.withHeaders {
				for name, values := range msg.Header() {
					name = textproto.CanonicalMIMEHeaderKey(name)
					for _, v := range values {
						attrs = append(attrs, attribute.String(fmt.Sprintf("nats.header.%s", name), v))
					}
				}
			}
			if cfg.withData {
				attrs = append(attrs, attribute.String("nats.data", string(msg.Data())))
			}
			if metadatable, ok := msg.(peanats.Metadatable); ok {
				if meta, err := metadatable.Metadata(); err == nil {
					attrs = append(attrs,
						attribute.Int64("nats.jetstream.stream_seq", int64(meta.Sequence.Stream)),
						attribute.Int64("nats.jetstream.consumer_seq", int64(meta.Sequence.Consumer)),
						attribute.Int("nats.jetstream.num_delivered", int(meta.NumDelivered)),
						attribute.Int("nats.jetstream.num_pending", int(meta.NumPending)),
						attribute.String("nats.jetstream.stream", meta.Stream),
						attribute.String("nats.jetstream.consumer", meta.Consumer),
						attribute.String("nats.jetstream.domain", meta.Domain),
					)
				}
			}
			propagator := otel.GetTextMapPropagator()
			ctxPropagator := propagator.Extract(ctx, propagation.HeaderCarrier(msg.Header()))
			ctxSpan, span := cfg.tracer.Start(ctxPropagator, cfg.spanName,
				trace.WithSpanKind(cfg.spanKind),
				trace.WithAttributes(attrs...),
			)
			defer span.End()
			err := next.HandleMsg(ctxSpan, msg)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return err
		})
	}
}
