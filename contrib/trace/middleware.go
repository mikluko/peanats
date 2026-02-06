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
	eventHeaders   bool
	eventData      bool
	truncateDataAt int
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

// MiddlewareWithEventHeaders enables adding a span event with message headers
func MiddlewareWithEventHeaders() MiddlewareOption {
	return func(o *middlewareOptions) {
		o.eventHeaders = true
	}
}

// MiddlewareWithEventData enables adding a span event with message data, truncated to the given length.
// A zero or negative truncateAt means no truncation.
func MiddlewareWithEventData(truncateAt int) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.eventData = true
		o.truncateDataAt = truncateAt
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
			spanAttrs := append(cfg.spanAttributes,
				attribute.String("nats.subject", msg.Subject()),
			)
			if metadatable, ok := msg.(peanats.Metadatable); ok {
				if meta, err := metadatable.Metadata(); err == nil {
					spanAttrs = append(spanAttrs,
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
			ctx = propagator.Extract(ctx, propagation.HeaderCarrier(msg.Header()))

			ctxSpan, span := cfg.tracer.Start(ctx, cfg.spanName,
				trace.WithSpanKind(cfg.spanKind),
				trace.WithAttributes(spanAttrs...),
			)
			defer span.End()

			if cfg.eventHeaders || cfg.eventData {
				var eventAttrs []attribute.KeyValue
				if cfg.eventHeaders {
					for name, values := range msg.Header() {
						name = textproto.CanonicalMIMEHeaderKey(name)
						for _, v := range values {
							eventAttrs = append(eventAttrs, attribute.String(fmt.Sprintf("nats.header.%s", name), v))
						}
					}
				}
				if cfg.eventData {
					dataFull := string(msg.Data())
					dataTrunc := dataFull
					if cfg.truncateDataAt > 0 && len(dataFull) > cfg.truncateDataAt {
						dataTrunc = dataFull[:cfg.truncateDataAt]
					}
					eventAttrs = append(eventAttrs,
						attribute.String("nats.data", dataTrunc),
						attribute.Int("nats.data_length", len(dataFull)),
						attribute.Bool("nats.data_truncated", len(dataFull) != len(dataTrunc)),
					)
				}
				span.AddEvent("nats.message", trace.WithAttributes(eventAttrs...))
			}

			err := next.HandleMsg(ctxSpan, msg)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return err
		})
	}
}
