package trace

import (
	"context"
	"net/textproto"
	"testing"

	"github.com/mikluko/peanats/publisher"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type mockPublisher struct {
	lastSubject string
	lastData    any
	lastOptions []publisher.PublishOption
}

func (m *mockPublisher) Publish(ctx context.Context, subject string, data any, opts ...publisher.PublishOption) error {
	m.lastSubject = subject
	m.lastData = data
	m.lastOptions = opts
	return nil
}

func TestTracingPublisher(t *testing.T) {
	// Setup test environment
	mock := &mockPublisher{}
	tracer := otel.Tracer("test")

	// Create tracing publisher
	pub := NewPublisher(mock, PublisherWithTracer(tracer))

	// Test publish with context
	ctx := context.Background()
	err := pub.Publish(ctx, "test.subject", "test data")

	// Verify
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if mock.lastSubject != "test.subject" {
		t.Errorf("Expected subject 'test.subject', got %s", mock.lastSubject)
	}

	// Check that headers were added
	var hasHeaders bool
	for _, opt := range mock.lastOptions {
		if opt != nil {
			hasHeaders = true
			break
		}
	}

	if !hasHeaders {
		t.Error("Expected headers to be added for trace propagation")
	}
}

func TestPublisherOptions(t *testing.T) {
	mock := &mockPublisher{}
	tracer := otel.Tracer("test")
	attrs := []attribute.KeyValue{
		attribute.String("service.name", "test"),
	}

	pub := NewPublisher(mock,
		PublisherWithTracer(tracer),
		PublisherWithAttributes(attrs...),
	)

	// Verify publisher was created
	if pub == nil {
		t.Fatal("NewPublisher returned nil")
	}

	// Test that it can publish
	ctx := context.Background()
	err := pub.Publish(ctx, "test", "data")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
}

func TestCanonicalizeHeader(t *testing.T) {
	t.Log(textproto.CanonicalMIMEHeaderKey("Traceparentid"))
}
