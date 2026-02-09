package trace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/textproto"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/requester"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// RequesterOption configures a trace-aware requester
type RequesterOption[RQ, RS any] func(*tracingRequester[RQ, RS])

// RequesterWithTracer sets the tracer for the requester
func RequesterWithTracer[RQ, RS any](tracer trace.Tracer) RequesterOption[RQ, RS] {
	return func(req *tracingRequester[RQ, RS]) {
		req.tracer = tracer
	}
}

// RequesterWithSpanName sets the span name for traces created by the requester
func RequesterWithSpanName[RQ, RS any](name string) RequesterOption[RQ, RS] {
	return func(req *tracingRequester[RQ, RS]) {
		req.spanName = name
	}
}

// RequesterWithAttributes adds attributes to all spans created by the requester
func RequesterWithAttributes[RQ, RS any](attrs ...attribute.KeyValue) RequesterOption[RQ, RS] {
	return func(req *tracingRequester[RQ, RS]) {
		req.attrs = append(req.attrs, attrs...)
	}
}

// RequesterWithEventHeaders enables adding message headers as span event attributes
func RequesterWithEventHeaders[RQ, RS any]() RequesterOption[RQ, RS] {
	return func(req *tracingRequester[RQ, RS]) {
		req.eventHeaders = true
	}
}

// RequesterWithEventData enables adding message data as span event attributes, truncated to the given length.
// A zero or negative truncateAt means no truncation.
func RequesterWithEventData[RQ, RS any](truncateAt int) RequesterOption[RQ, RS] {
	return func(req *tracingRequester[RQ, RS]) {
		req.eventData = true
		req.truncateDataAt = truncateAt
	}
}

type tracingRequester[RQ, RS any] struct {
	requester.Requester[RQ, RS]
	tracer         trace.Tracer
	spanName       string
	attrs          []attribute.KeyValue
	eventHeaders   bool
	eventData      bool
	truncateDataAt int
}

// NewRequester creates a new trace-aware requester that implements requester.Requester
func NewRequester[RQ, RS any](req requester.Requester[RQ, RS], opts ...RequesterOption[RQ, RS]) requester.Requester[RQ, RS] {
	res := &tracingRequester[RQ, RS]{
		Requester: req,
		tracer:    otel.Tracer("peanats"),
		spanName:  "peanats.request",
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

// buildMessageEventAttrs builds span event attributes from headers and data.
func buildMessageEventAttrs(header peanats.Header, data any, eventHeaders, eventData bool, truncateAt int) []attribute.KeyValue {
	var attrs []attribute.KeyValue
	if eventHeaders && header != nil {
		for name, values := range header {
			name = textproto.CanonicalMIMEHeaderKey(name)
			for _, v := range values {
				attrs = append(attrs, attribute.String(fmt.Sprintf("nats.header.%s", name), v))
			}
		}
	}
	if eventData && data != nil {
		b, err := json.Marshal(data)
		if err == nil {
			dataFull := string(b)
			dataTrunc := dataFull
			if truncateAt > 0 && len(dataFull) > truncateAt {
				dataTrunc = dataFull[:truncateAt]
			}
			attrs = append(attrs,
				attribute.String("nats.data", dataTrunc),
				attribute.Int("nats.data_length", len(dataFull)),
				attribute.Bool("nats.data_truncated", len(dataFull) != len(dataTrunc)),
			)
		}
	}
	return attrs
}

// Request sends a request with trace context propagation
func (r *tracingRequester[RQ, RS]) Request(ctx context.Context, subject string, data *RQ, opts ...requester.RequestOption) (requester.Response[RS], error) {
	// Start a new span for the request operation
	spanOpts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(append(r.attrs, attribute.String("nats.subject", subject))...),
	}
	ctx, span := r.tracer.Start(ctx, r.spanName, spanOpts...)
	defer span.End()

	// Inject trace context into the request headers
	header := make(peanats.Header)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))

	// Emit request event
	if attrs := buildMessageEventAttrs(header, data, r.eventHeaders, r.eventData, r.truncateDataAt); len(attrs) > 0 {
		span.AddEvent("nats.request", trace.WithAttributes(attrs...))
	}

	// Execute the request with trace headers
	resp, err := r.Requester.Request(ctx, subject, data, append(opts, requester.RequestHeader(header))...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Emit response event
	if attrs := buildMessageEventAttrs(resp.Header(), resp.Value(), r.eventHeaders, r.eventData, r.truncateDataAt); len(attrs) > 0 {
		span.AddEvent("nats.response", trace.WithAttributes(attrs...))
	}

	return resp, nil
}

// ResponseReceiver creates a response receiver with trace context propagation
func (r *tracingRequester[RQ, RS]) ResponseReceiver(ctx context.Context, subject string, data *RQ, opts ...requester.ResponseReceiverOption) (requester.ResponseReceiver[RS], error) {
	// Start a new span for the response receiver operation
	spanOpts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(append(r.attrs, attribute.String("nats.subject", subject))...),
	}
	ctx, span := r.tracer.Start(ctx, r.spanName, spanOpts...)

	// Inject trace context into the request headers
	header := make(peanats.Header)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))

	// Emit request event
	if attrs := buildMessageEventAttrs(header, data, r.eventHeaders, r.eventData, r.truncateDataAt); len(attrs) > 0 {
		span.AddEvent("nats.request", trace.WithAttributes(attrs...))
	}

	// Create response receiver with trace headers
	reqOpts := []requester.RequestOption{requester.RequestHeader(header)}
	receiver, err := r.Requester.ResponseReceiver(ctx, subject, data, append(opts, requester.ResponseReceiverRequestOptions(reqOpts...))...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, err
	}

	return &tracingResponseReceiver[RS]{
		ResponseReceiver: receiver,
		span:             span,
		tracer:           r.tracer,
		eventHeaders:     r.eventHeaders,
		eventData:        r.eventData,
		truncateDataAt:   r.truncateDataAt,
	}, nil
}

type tracingResponseReceiver[T any] struct {
	requester.ResponseReceiver[T]
	span           trace.Span
	tracer         trace.Tracer
	eventHeaders   bool
	eventData      bool
	truncateDataAt int
}

// Next retrieves the next response with trace context extraction
func (r *tracingResponseReceiver[T]) Next(ctx context.Context) (requester.Response[T], error) {
	resp, err := r.ResponseReceiver.Next(ctx)
	if err != nil {
		// Record error and return
		if !errors.Is(err, requester.ErrSkip) && !errors.Is(err, requester.ErrOver) {
			r.span.RecordError(err)
		}
		return nil, err
	}

	// Extract trace context from response headers if available
	if resp.Header() != nil {
		propagator := otel.GetTextMapPropagator()
		_ = propagator.Extract(ctx, propagation.HeaderCarrier(resp.Header()))
	}

	// Emit response event
	if attrs := buildMessageEventAttrs(resp.Header(), resp.Value(), r.eventHeaders, r.eventData, r.truncateDataAt); len(attrs) > 0 {
		r.span.AddEvent("nats.response", trace.WithAttributes(attrs...))
	}

	return resp, nil
}

// Stop stops the response receiver and ends the trace span
func (r *tracingResponseReceiver[T]) Stop() error {
	defer r.span.End()
	err := r.ResponseReceiver.Stop()
	if err != nil {
		r.span.RecordError(err)
		r.span.SetStatus(codes.Error, err.Error())
	}
	return err
}
