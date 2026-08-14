package trace

import (
	"context"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/mikluko/peanats"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
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
		MiddlewareWithEventHeaders(),
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
		MiddlewareWithEventHeaders(),
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

// messageEventAttrs extracts the nats.message span event attributes from the
// single exported span as a key-indexed map.
func messageEventAttrs(t *testing.T, exporter *tracetest.InMemoryExporter) map[string]attribute.Value {
	t.Helper()
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 exported span, got %d", len(spans))
	}
	for _, ev := range spans[0].Events {
		if ev.Name == "nats.message" {
			attrs := make(map[string]attribute.Value, len(ev.Attributes))
			for _, kv := range ev.Attributes {
				attrs[string(kv.Key)] = kv.Value
			}
			return attrs
		}
	}
	t.Fatal("nats.message span event not found")
	return nil
}

func handleWithEventData(t *testing.T, msg *mockMsg, truncateAt int) *tracetest.InMemoryExporter {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	middleware := Middleware(
		MiddlewareWithTracer(tp.Tracer("test")),
		MiddlewareWithSpanKind(trace.SpanKindConsumer),
		MiddlewareWithEventHeaders(),
		MiddlewareWithEventData(truncateAt),
	)
	if err := middleware(&mockHandler{}).HandleMsg(context.Background(), msg); err != nil {
		t.Fatalf("HandleMsg failed: %v", err)
	}
	return exporter
}

func TestMiddleware_EventDataTextPayload(t *testing.T) {
	exporter := handleWithEventData(t, &mockMsg{
		subject: "test.subject",
		data:    []byte("hello world"),
		header:  make(peanats.Header),
	}, 1024)

	attrs := messageEventAttrs(t, exporter)
	assert.Equal(t, "hello world", attrs["nats.data"].AsString())
	assert.Equal(t, int64(11), attrs["nats.data_length"].AsInt64())
	assert.False(t, attrs["nats.data_truncated"].AsBool())
	_, hasEncoding := attrs["nats.data_encoding"]
	assert.False(t, hasEncoding)
}

func TestMiddleware_EventDataBinaryPayload(t *testing.T) {
	exporter := handleWithEventData(t, &mockMsg{
		subject: "test.subject",
		data:    []byte{0xff, 0xfe, 0x41, 0x80},
		header:  make(peanats.Header),
	}, 1024)

	attrs := messageEventAttrs(t, exporter)
	_, hasData := attrs["nats.data"]
	assert.False(t, hasData, "binary payload must not be recorded as nats.data")
	assert.Equal(t, "binary", attrs["nats.data_encoding"].AsString())
	assert.Equal(t, int64(4), attrs["nats.data_length"].AsInt64())
	assert.False(t, attrs["nats.data_truncated"].AsBool())
}

func TestMiddleware_EventDataTruncationKeepsRuneBoundary(t *testing.T) {
	// 10 ascii bytes, then a 3-byte euro sign, then 3 more bytes: truncation at
	// 12 lands inside the euro sign and must back off to the rune boundary.
	payload := append([]byte("aaaaaaaaaa"), []byte("€zzz")...)
	exporter := handleWithEventData(t, &mockMsg{
		subject: "test.subject",
		data:    payload,
		header:  make(peanats.Header),
	}, 12)

	attrs := messageEventAttrs(t, exporter)
	assert.Equal(t, "aaaaaaaaaa", attrs["nats.data"].AsString())
	assert.Equal(t, int64(16), attrs["nats.data_length"].AsInt64())
	assert.True(t, attrs["nats.data_truncated"].AsBool())
	_, hasEncoding := attrs["nats.data_encoding"]
	assert.False(t, hasEncoding)
}

func TestMiddleware_EventHeadersInvalidUTF8Skipped(t *testing.T) {
	header := make(peanats.Header)
	header.Set("X-Good", "ok")
	header["X-Bad"] = []string{"\xff\xfe"}

	exporter := handleWithEventData(t, &mockMsg{
		subject: "test.subject",
		data:    []byte("test data"),
		header:  header,
	}, 1024)

	attrs := messageEventAttrs(t, exporter)
	assert.Equal(t, "ok", attrs["nats.header.X-Good"].AsString())
	_, hasBad := attrs["nats.header.X-Bad"]
	assert.False(t, hasBad, "invalid UTF-8 header values must be skipped")
}

func TestMiddleware_ConcurrentHandlersShareNoSpanAttributeBacking(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	// Three separate option calls leave the accumulated slice with spare
	// capacity, so per-message appends onto a shared backing array would be
	// concurrent writes to the same element. Run under -race.
	middleware := Middleware(
		MiddlewareWithTracer(tp.Tracer("test")),
		MiddlewareWithSpanAttributes(attribute.String("a", "1")),
		MiddlewareWithSpanAttributes(attribute.String("b", "2")),
		MiddlewareWithSpanAttributes(attribute.String("c", "3")),
	)
	wrapped := middleware(peanats.MsgHandlerFunc(func(context.Context, peanats.Msg) error {
		return nil
	}))

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := &mockMsg{subject: "test.subject", data: []byte("x"), header: make(peanats.Header)}
			assert.NoError(t, wrapped.HandleMsg(context.Background(), msg))
		}()
	}
	wg.Wait()
}

func TestBuildMessageEventAttrs_JSONTruncationStaysValidUTF8(t *testing.T) {
	data := map[string]string{"msg": "aaaaaaaaaa€zzzzzzzzzz"}
	attrs := buildMessageEventAttrs(nil, data, false, true, 16)

	var got *attribute.Value
	for _, kv := range attrs {
		if string(kv.Key) == "nats.data" {
			v := kv.Value
			got = &v
		}
		assert.NotEqual(t, "nats.data_encoding", string(kv.Key), "JSON payloads are always valid UTF-8")
	}
	if got == nil {
		t.Fatal("nats.data attribute not found")
	}
	assert.True(t, utf8.ValidString(got.AsString()), "truncated JSON must remain valid UTF-8")
}
