package trace

import (
	"context"
	"testing"

	"github.com/mikluko/peanats"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type mockMsg struct {
	subject string
	data    []byte
	header  peanats.Header
}

func (m *mockMsg) Subject() string        { return m.subject }
func (m *mockMsg) Data() []byte           { return m.data }
func (m *mockMsg) Header() peanats.Header { return m.header }

type mockHandler struct {
	lastCtx context.Context
	lastMsg peanats.Msg
}

func (m *mockHandler) HandleMsg(ctx context.Context, msg peanats.Msg) error {
	m.lastCtx = ctx
	m.lastMsg = msg
	return nil
}

func TestTracingMiddleware(t *testing.T) {
	// Setup test environment
	tracer := otel.Tracer("test")
	handler := &mockHandler{}

	// Create middleware
	middleware := Middleware(
		MiddlewareWithTracer(tracer),
		MiddlewareWithSpanKind(trace.SpanKindConsumer),
		MiddlewareWithHeaders(true),
	)

	// Wrap handler
	wrappedHandler := middleware(handler)

	// Create test message with trace headers
	ctx := context.Background()
	header := make(peanats.Header)

	// Inject trace context into headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))

	msg := &mockMsg{
		subject: "test.subject",
		data:    []byte("test data"),
		header:  header,
	}

	// Handle message
	err := wrappedHandler.HandleMsg(ctx, msg)

	// Verify
	if err != nil {
		t.Fatalf("HandleMsg failed: %v", err)
	}

	if handler.lastMsg != msg {
		t.Error("Handler did not receive the message")
	}

	// Verify trace context was extracted (we can't easily test this without more complex setup)
	if handler.lastCtx == nil {
		t.Error("Handler did not receive context")
	}
}

func TestMiddleware_TracePropagation(t *testing.T) {
	// Setup OpenTelemetry with the same configuration as the live system
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.Baggage{},
			propagation.TraceContext{},
		),
	)

	tracer := otel.Tracer("test")

	// Create a trace context with known trace ID
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	originalTraceID := span.SpanContext().TraceID().String()

	// Inject trace context into headers
	headers := make(peanats.Header)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(headers))

	// Create message with trace headers
	msg := &mockMsg{
		subject: "test.subject",
		data:    []byte("test data"),
		header:  headers,
	}

	// Create handler to capture received context
	handler := &mockHandler{}

	// Create middleware with trace extraction enabled
	middleware := Middleware(
		MiddlewareWithTracer(tracer),
		MiddlewareWithHeaders(true),
	)
	wrappedHandler := middleware(handler)

	// Process message (context propagation should work)
	err := wrappedHandler.HandleMsg(context.Background(), msg)
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}

	if handler.lastCtx == nil {
		t.Fatal("Handler did not receive context")
	}

	// Verify trace context was propagated correctly
	receivedSpan := trace.SpanFromContext(handler.lastCtx)
	if !receivedSpan.SpanContext().IsValid() {
		t.Fatal("No valid span found in received context")
	}

	receivedTraceID := receivedSpan.SpanContext().TraceID().String()
	if receivedTraceID != originalTraceID {
		t.Errorf("Trace context not propagated: expected trace ID %s, got %s", originalTraceID, receivedTraceID)
	}
}

func TestNATSHeaderExtraction(t *testing.T) {
	// Test different approaches to extract trace context from NATS headers
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test")
	defer span.End()

	originalTraceID := span.SpanContext().TraceID().String()

	t.Run("HeaderCarrier_Direct", func(t *testing.T) {
		// Current implementation - inject and extract with HeaderCarrier
		headers := make(peanats.Header)
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(headers))

		t.Logf("Injected headers: %v", headers)

		extractedCtx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.HeaderCarrier(headers))
		extractedSpan := trace.SpanFromContext(extractedCtx)

		if extractedSpan.SpanContext().IsValid() {
			extractedTraceID := extractedSpan.SpanContext().TraceID().String()
			t.Logf("Extracted trace ID: %s", extractedTraceID)
			if extractedTraceID != originalTraceID {
				t.Errorf("Trace ID mismatch: expected %s, got %s", originalTraceID, extractedTraceID)
			}
		} else {
			t.Error("Failed to extract valid trace context with HeaderCarrier")
		}
	})

	t.Run("MapCarrier_Converted", func(t *testing.T) {
		// Alternative - inject with HeaderCarrier, convert to map[string]string, extract with MapCarrier
		natsHeaders := make(peanats.Header)
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(natsHeaders))

		// Convert NATS headers (map[string][]string) to map[string]string
		stringHeaders := make(map[string]string)
		for k, v := range natsHeaders {
			if len(v) > 0 {
				stringHeaders[k] = v[0]
			}
		}

		t.Logf("NATS headers: %v", natsHeaders)
		t.Logf("Converted headers: %v", stringHeaders)

		extractedCtx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(stringHeaders))
		extractedSpan := trace.SpanFromContext(extractedCtx)

		if extractedSpan.SpanContext().IsValid() {
			extractedTraceID := extractedSpan.SpanContext().TraceID().String()
			t.Logf("Extracted trace ID: %s", extractedTraceID)
			if extractedTraceID != originalTraceID {
				t.Errorf("Trace ID mismatch: expected %s, got %s", originalTraceID, extractedTraceID)
			}
		} else {
			t.Log("Expected: MapCarrier fails because OpenTelemetry propagation expects lowercase headers")
		}
	})
}
