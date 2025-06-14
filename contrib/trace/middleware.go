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
	attributes     []attribute.KeyValue
	extractHeaders bool
	recordData     bool
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

// MiddlewareWithAttributes adds attributes to all spans created by the middleware
func MiddlewareWithAttributes(attrs ...attribute.KeyValue) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.attributes = append(o.attributes, attrs...)
	}
}

// MiddlewareWithHeaders enables trace context extraction from message headers
func MiddlewareWithHeaders(extract bool) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.extractHeaders = extract
	}
}

// MiddlewareWithPayload enables recording message data as span attributes
func MiddlewareWithPayload(record bool) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.recordData = record
	}
}

// Middleware creates a NATS message middleware that handles OpenTelemetry trace propagation
func Middleware(opts ...MiddlewareOption) peanats.MsgMiddleware {
	options := &middlewareOptions{
		tracer:         otel.Tracer("peanats"),
		spanKind:       trace.SpanKindUnspecified,
		extractHeaders: true,
		recordData:     false,
	}
	for _, opt := range opts {
		opt(options)
	}

	return func(next peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
			propagator := otel.GetTextMapPropagator()
			ctx = propagator.Extract(ctx, propagation.HeaderCarrier(msg.Header()))

			checkSpan := trace.SpanFromContext(ctx)
			_ = checkSpan // Ensure span is initialized

			attrs := append(options.attributes,
				attribute.String("nats.subject", msg.Subject()),
			)
			for name, values := range msg.Header() {
				name = textproto.CanonicalMIMEHeaderKey(name)
				for _, v := range values {
					attrs = append(attrs, attribute.String(fmt.Sprintf("nats.header.%s", name), v))
				}
			}
			if options.recordData {
				attrs = append(attrs, attribute.String("nats.payload", string(msg.Data())))
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
			ctx, span := options.tracer.Start(ctx, "receive",
				trace.WithSpanKind(options.spanKind),
				trace.WithAttributes(attrs...),
			)
			defer span.End()
			err := next.HandleMsg(ctx, msg)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return err
		})
	}
}
