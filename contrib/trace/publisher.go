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
type PublisherOption func(*publisherOptions)

type publisherOptions struct {
	tracer     trace.Tracer
	attributes []attribute.KeyValue
}

// PublisherWithTracer sets the tracer for the publisher
func PublisherWithTracer(tracer trace.Tracer) PublisherOption {
	return func(o *publisherOptions) {
		o.tracer = tracer
	}
}

// PublisherWithAttributes adds attributes to all spans created by the publisher
func PublisherWithAttributes(attrs ...attribute.KeyValue) PublisherOption {
	return func(o *publisherOptions) {
		o.attributes = append(o.attributes, attrs...)
	}
}

type tracingPublisher struct {
	publisher.Publisher
	tracer     trace.Tracer
	attributes []attribute.KeyValue
}

// NewPublisher creates a new trace-aware publisher that implements publisher.Publisher
func NewPublisher(pub publisher.Publisher, opts ...PublisherOption) publisher.Publisher {
	options := &publisherOptions{
		tracer: otel.Tracer("peanats"),
	}
	for _, opt := range opts {
		opt(options)
	}
	return &tracingPublisher{
		Publisher:  pub,
		tracer:     options.tracer,
		attributes: options.attributes,
	}
}

// Publish publishes a message with trace context propagation
func (p *tracingPublisher) Publish(ctx context.Context, subject string, data any, opts ...publisher.PublishOption) error {
	// Start a new span for the publish operation
	attrs := append(p.attributes,
		attribute.String("nats.subject", subject),
	)
	ctx, span := p.tracer.Start(ctx, "publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attrs...),
	)
	defer span.End()

	header := make(peanats.Header)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))

	opts = append(opts, publisher.WithHeader(header))

	err := p.Publisher.Publish(ctx, subject, data, opts...)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
