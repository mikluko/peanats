package otel

import (
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel/attribute"

	"github.com/mikluko/peanats"
)

func MessageSubjectAttribute(subj string) attribute.KeyValue {
	return attribute.String("nats.subject", subj)
}

func MessageHeaderAttributes(header peanats.Header) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(header))
	for k := range header {
		if v := header.Values(k); len(v) > 1 {
			attrs = append(attrs, attribute.StringSlice("nats.header."+k, v))
		} else {
			attrs = append(attrs, attribute.String("nats.header."+k, header.Get(k)))
		}
	}
	return attrs
}

func MessageDataAttribute(data []byte) attribute.KeyValue {
	return attribute.String("nats.data", string(data))
}

func MessageMetadataAttributes(meta *jetstream.MsgMetadata) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("nats.meta.stream", meta.Stream),
		attribute.String("nats.meta.consumer", meta.Consumer),
		attribute.String("nats.meta.timestamp", meta.Timestamp.Format(time.RFC3339)),
		attribute.Int64("nats.meta.sequence.stream", int64(meta.Sequence.Stream)),
		attribute.Int64("nats.meta.sequence.consumer", int64(meta.Sequence.Consumer)),
		attribute.Int64("nats.meta.num_delivered", int64(meta.NumDelivered)),
		attribute.Int64("nats.meta.num_pending", int64(meta.NumPending)),
	}
}
