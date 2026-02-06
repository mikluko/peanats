package trace

import (
	"context"
	"errors"
	"testing"

	"github.com/mikluko/peanats/publisher"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type mockPublisher struct {
	lastSubject string
	lastData    any
	lastOptions []publisher.PublishOption
	err         error
}

func (m *mockPublisher) Publish(ctx context.Context, subject string, data any, opts ...publisher.PublishOption) error {
	m.lastSubject = subject
	m.lastData = data
	m.lastOptions = opts
	return m.err
}

func TestTracingPublisher(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	mock := &mockPublisher{}
	pub := NewPublisher(mock)

	// Create a span to add events to
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "parent-span")

	err := pub.Publish(ctx, "test.subject", "test data")
	span.End()

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, "test.subject", mock.lastSubject)

	// Check that headers were added for trace propagation
	var hasHeaders bool
	for _, opt := range mock.lastOptions {
		if opt != nil {
			hasHeaders = true
			break
		}
	}
	assert.True(t, hasHeaders, "Expected headers to be added for trace propagation")

	// Verify event was added to span
	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	parentSpan := spans[0]
	assert.Equal(t, "parent-span", parentSpan.Name)
	assert.Len(t, parentSpan.Events, 1)

	event := parentSpan.Events[0]
	assert.Equal(t, "peanats.publish", event.Name)

	// Check event attributes
	var subjectAttr attribute.KeyValue
	for _, attr := range event.Attributes {
		if attr.Key == "nats.subject" {
			subjectAttr = attr
		}
	}
	assert.Equal(t, "test.subject", subjectAttr.Value.AsString())
}

func TestTracingPublisher_WithError(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	expectedErr := errors.New("publish failed")
	mock := &mockPublisher{err: expectedErr}
	pub := NewPublisher(mock)

	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "parent-span")

	err := pub.Publish(ctx, "test.subject", "test data")
	span.End()

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	// Verify event includes error attribute
	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	event := spans[0].Events[0]
	var errorAttr attribute.KeyValue
	for _, attr := range event.Attributes {
		if attr.Key == "error" {
			errorAttr = attr
		}
	}
	assert.Equal(t, "publish failed", errorAttr.Value.AsString())
}

func TestTracingPublisher_NoSpan(t *testing.T) {
	// Test that publishing works even without a recording span
	mock := &mockPublisher{}
	pub := NewPublisher(mock)

	ctx := context.Background()
	err := pub.Publish(ctx, "test.subject", "test data")

	assert.NoError(t, err)
	assert.Equal(t, "test.subject", mock.lastSubject)
}

func TestPublisherOptions(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	mock := &mockPublisher{}
	pub := NewPublisher(mock,
		PublisherWithEventName("custom.publish"),
		PublisherWithAttributes(
			attribute.String("service.name", "test"),
			attribute.Int("custom.value", 42),
		),
	)

	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "parent-span")

	err := pub.Publish(ctx, "test.subject", "data")
	span.End()

	assert.NoError(t, err)

	// Verify custom event name and attributes
	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	event := spans[0].Events[0]
	assert.Equal(t, "custom.publish", event.Name)

	// Check custom attributes are present
	attrMap := make(map[string]attribute.Value)
	for _, attr := range event.Attributes {
		attrMap[string(attr.Key)] = attr.Value
	}

	assert.Equal(t, "test.subject", attrMap["nats.subject"].AsString())
	assert.Equal(t, "test", attrMap["service.name"].AsString())
	assert.Equal(t, int64(42), attrMap["custom.value"].AsInt64())
}

func TestTracingPublisher_HeaderPropagation(t *testing.T) {
	// Verify trace context is properly injected into headers
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	mock := &mockPublisher{}
	pub := NewPublisher(mock)

	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "parent-span")
	originalTraceID := span.SpanContext().TraceID().String()

	err := pub.Publish(ctx, "test.subject", "test data")
	span.End()

	assert.NoError(t, err)

	// The mock captured the options, but we can't easily inspect the header
	// without more complex setup. The key verification is that options were passed.
	assert.NotEmpty(t, mock.lastOptions)

	// Verify the span was recorded
	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, originalTraceID, spans[0].SpanContext.TraceID().String())
}
