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

// PublisherWithTracer sets the tracer for the publisher
func PublisherWithTracer(tracer trace.Tracer) PublisherOption {
	return func(pub *tracingPublisher) {
		pub.tracer = tracer
	}
}

// PublisherWithAttributes adds attributes to all spans created by the publisher
func PublisherWithAttributes(attrs ...attribute.KeyValue) PublisherOption {
	return func(pub *tracingPublisher) {
		pub.opts = append(pub.opts, trace.WithAttributes(attrs...))
	}
}

// PublisherWithNewRoot indicates that the publisher should start a new root span for each publish operation
func PublisherWithNewRoot() PublisherOption {
	return func(pub *tracingPublisher) {
		pub.root = true
	}
}

// PublisherWithLinks adds a links to the span created by the publisher
func PublisherWithLinks(links ...trace.Link) PublisherOption {
	return func(pub *tracingPublisher) {
		pub.opts = append(pub.opts, trace.WithLinks(links...))
	}
}

type tracingPublisher struct {
	publisher.Publisher
	tracer trace.Tracer
	opts   []trace.SpanStartOption
	root   bool // indicates if a new root span should be created for each publish operation
}

// NewPublisher creates a new trace-aware publisher that implements publisher.Publisher
func NewPublisher(pub publisher.Publisher, opts ...PublisherOption) publisher.Publisher {
	res := tracingPublisher{
		Publisher: pub,
	}
	for _, opt := range opts {
		opt(&res)
	}
	return &res
}

// Publish publishes a message with trace context propagation
func (p *tracingPublisher) Publish(ctx context.Context, subject string, data any, opts ...publisher.PublishOption) error {
	// Start a new span for the publish operation
	spanOpts := append(p.opts,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attribute.String("nats.subject", subject)),
	)
	if p.root {
		spanOpts = append(spanOpts, trace.WithNewRoot(), trace.WithLinks(trace.LinkFromContext(ctx)))
	}
	ctx, span := p.tracer.Start(ctx, "publish", spanOpts...)
	defer span.End()

	header := make(peanats.Header)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))

	err := p.Publisher.Publish(ctx, subject, data, append(opts, publisher.WithHeader(header))...)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
