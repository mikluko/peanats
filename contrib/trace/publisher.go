package trace

import (
	"context"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/publisher"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// PublisherOption configures a trace-aware publisher
type PublisherOption func(*tracingPublisher)

// PublisherWithEventName sets the event name for publish operations
func PublisherWithEventName(name string) PublisherOption {
	return func(pub *tracingPublisher) {
		pub.eventName = name
	}
}

// PublisherWithAttributes adds attributes to all events created by the publisher
func PublisherWithAttributes(attrs ...attribute.KeyValue) PublisherOption {
	return func(pub *tracingPublisher) {
		pub.attrs = append(pub.attrs, attrs...)
	}
}

type tracingPublisher struct {
	publisher.Publisher
	eventName string
	attrs     []attribute.KeyValue
}

// NewPublisher creates a new trace-aware publisher that implements publisher.Publisher
func NewPublisher(pub publisher.Publisher, opts ...PublisherOption) publisher.Publisher {
	res := tracingPublisher{
		Publisher: pub,
		eventName: "peanats.publish",
	}
	for _, opt := range opts {
		opt(&res)
	}
	return &res
}

// Publish publishes a message with trace context propagation
func (p *tracingPublisher) Publish(ctx context.Context, subject string, data any, opts ...publisher.PublishOption) error {
	span := trace.SpanFromContext(ctx)

	// Inject trace context into headers for propagation
	header := make(peanats.Header)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))

	err := p.Publisher.Publish(ctx, subject, data, append(opts, publisher.WithHeader(header))...)

	// Add event to span if it's recording
	if span.IsRecording() {
		attrs := append([]attribute.KeyValue{attribute.String("nats.subject", subject)}, p.attrs...)
		if err != nil {
			attrs = append(attrs, attribute.String("error", err.Error()))
		}
		span.AddEvent(p.eventName, trace.WithAttributes(attrs...))
	}

	return err
}
