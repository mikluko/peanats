package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	peapublisher "github.com/mikluko/peanats/publisher"
)

type PublisherOption func(*publisherParams)

func PublisherTracer(t trace.Tracer) PublisherOption {
	return func(p *publisherParams) {
		p.tracer = t
	}
}

func PublisherAttributes(attrs []attribute.KeyValue) PublisherOption {
	return func(p *publisherParams) {
		p.attrs = attrs
	}
}

type publisherParams struct {
	tracer trace.Tracer
	attrs  []attribute.KeyValue
}

func Publisher(pub peapublisher.Publisher, opts ...PublisherOption) peapublisher.Publisher {
	p := publisherParams{
		tracer: otel.Tracer("peanats/publisher"),
	}
	for _, opt := range opts {
		opt(&p)
	}
	return &publisherImpl{pub, p}
}

type publisherImpl struct {
	peapublisher.Publisher
	params publisherParams
}

func (p *publisherImpl) Publish(ctx context.Context, subj string, x any, opts ...peapublisher.PublishOption) error {
	attrs := append(p.params.attrs, MessageSubjectAttribute(subj))
	ctx, span := p.params.tracer.Start(ctx, "peanats/publisher",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(attrs...),
	)
	defer span.End()
	return p.Publisher.Publish(ctx, subj, x, opts...)
}
