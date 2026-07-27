package trace

import (
	"context"
	"fmt"
	"net/textproto"
	"unicode/utf8"

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
// A zero or negative truncateAt means no truncation. The payload is recorded as the nats.data
// attribute only when it is valid UTF-8; binary payloads are marked with nats.data_encoding=binary
// instead, keeping the span exportable.
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

			var eventAttrs []attribute.KeyValue
			if metadatable, ok := msg.(peanats.Metadatable); ok {
				if meta, err := metadatable.Metadata(); err == nil {
					eventAttrs = append(eventAttrs,
						attribute.Int64("nats.jetstream.stream_seq", int64(meta.Sequence.Stream)),
						attribute.Int64("nats.jetstream.consumer_seq", int64(meta.Sequence.Consumer)),
						attribute.Int("nats.jetstream.num_delivered", int(meta.NumDelivered)),
						attribute.Int("nats.jetstream.num_pending", int(meta.NumPending)),
					)
				}
			}
			if cfg.eventHeaders {
				eventAttrs = appendMessageHeaderEventAttrs(eventAttrs, msg.Header())
			}
			if cfg.eventData {
				eventAttrs = appendMessageDataEventAttrs(eventAttrs, msg.Data(), cfg.truncateDataAt)
			}
			if len(eventAttrs) > 0 {
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

// appendMessageHeaderEventAttrs appends nats.header.* attributes for each
// header value. Headers carrying invalid UTF-8 in their name or value are
// skipped: OTLP string fields must be valid UTF-8, and a single offending
// value fails marshaling of the whole span batch on export.
func appendMessageHeaderEventAttrs(attrs []attribute.KeyValue, header peanats.Header) []attribute.KeyValue {
	for name, values := range header {
		name = textproto.CanonicalMIMEHeaderKey(name)
		if !utf8.ValidString(name) {
			continue
		}
		for _, v := range values {
			if !utf8.ValidString(v) {
				continue
			}
			attrs = append(attrs, attribute.String(fmt.Sprintf("nats.header.%s", name), v))
		}
	}
	return attrs
}

// appendMessageDataEventAttrs appends message data attributes: payload length
// and truncation flag always; the payload itself only when it is valid UTF-8.
// Binary payloads (e.g. protobuf) are marked with nats.data_encoding=binary
// instead — a raw binary string would be invalid UTF-8 and fail OTLP
// marshaling, dropping the whole span batch on export.
func appendMessageDataEventAttrs(attrs []attribute.KeyValue, dataFull []byte, truncateAt int) []attribute.KeyValue {
	dataTrunc := dataFull
	truncated := false
	if truncateAt > 0 && len(dataFull) > truncateAt {
		dataTrunc = dataFull[:truncateAt]
		truncated = true
		// Back off to a rune boundary so truncation never splits a multi-byte
		// UTF-8 rune and turns a text payload "binary".
		for !utf8.Valid(dataTrunc) && utf8.Valid(dataFull) {
			dataTrunc = dataTrunc[:len(dataTrunc)-1]
		}
	}
	attrs = append(attrs,
		attribute.Int("nats.data_length", len(dataFull)),
		attribute.Bool("nats.data_truncated", truncated),
	)
	if utf8.Valid(dataTrunc) {
		attrs = append(attrs, attribute.String("nats.data", string(dataTrunc)))
	} else {
		attrs = append(attrs, attribute.String("nats.data_encoding", "binary"))
	}
	return attrs
}
