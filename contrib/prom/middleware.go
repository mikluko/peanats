package prom

import (
	"context"
	"time"

	"github.com/mikluko/peanats"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Option configures Prometheus middleware
type Option func(*params)

type params struct {
	namespace  string
	subsystem  string
	registerer prometheus.Registerer
}

// MiddlewareNamespace sets the Prometheus namespace for metrics
func MiddlewareNamespace(namespace string) Option {
	return func(p *params) {
		p.namespace = namespace
	}
}

// MiddlewareSubsystem sets the Prometheus subsystem for metrics
func MiddlewareSubsystem(subsystem string) Option {
	return func(p *params) {
		p.subsystem = subsystem
	}
}

// MiddlewareRegisterer sets a custom Prometheus registerer (for testing)
func MiddlewareRegisterer(registerer prometheus.Registerer) Option {
	return func(p *params) {
		p.registerer = registerer
	}
}

// Middleware creates a peanats.MsgMiddleware that records message processing metrics
func Middleware(opts ...Option) peanats.MsgMiddleware {
	// Apply default configuration
	p := params{
		namespace:  "peanats",
		subsystem:  "",
		registerer: prometheus.DefaultRegisterer,
	}
	for _, opt := range opts {
		opt(&p)
	}

	// Total number of messages processed by subject and status
	msgsProcessed := promauto.With(p.registerer).NewCounterVec(prometheus.CounterOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      "processed_total",
		Help:      "Total number of messages processed",
	}, []string{"subject", "status"})

	// Message processing latency histogram by subject
	msgDuration := promauto.With(p.registerer).NewHistogramVec(prometheus.HistogramOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      "latency_seconds",
		Help:      "Latency of message processing in seconds",
		Buckets:   prometheus.DefBuckets,
	}, []string{"subject"})

	// Messages currently being processed
	msgsInFlight := promauto.With(p.registerer).NewGaugeVec(prometheus.GaugeOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      "in_flight",
		Help:      "Number of messages currently being processed",
	}, []string{"subject"})

	// Message acknowledgment counters by subject and ack type
	msgsAcked := promauto.With(p.registerer).NewCounterVec(prometheus.CounterOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      "acked_total",
		Help:      "Total number of message acknowledgments",
	}, []string{"subject", "type"})

	return func(next peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
			subject := msg.Subject()

			// Wrap ackable messages to track acknowledgment metrics
			if _, ok := msg.(peanats.Ackable); ok {
				msg = &ackableWrapper{
					Msg:     msg,
					counter: msgsAcked,
				}
			}

			// Track in-flight messages
			msgsInFlight.WithLabelValues(subject).Inc()
			defer msgsInFlight.WithLabelValues(subject).Dec()

			// Track processing duration
			start := time.Now()
			defer func() {
				msgDuration.WithLabelValues(subject).Observe(time.Since(start).Seconds())
			}()

			// Process the message
			err := next.HandleMsg(ctx, msg)

			// Track message count with status
			status := "success"
			if err != nil {
				status = "error"
			}
			msgsProcessed.WithLabelValues(subject, status).Inc()

			return err
		})
	}
}

// ackableWrapper wraps an Ackable message to track acknowledgment metrics
type ackableWrapper struct {
	peanats.Msg
	counter *prometheus.CounterVec
}

func (a *ackableWrapper) Ack(ctx context.Context) error {
	err := a.Msg.(peanats.Ackable).Ack(ctx)
	a.counter.WithLabelValues(a.Subject(), "ack").Inc()
	return err
}

func (a *ackableWrapper) Nak(ctx context.Context) error {
	err := a.Msg.(peanats.Ackable).Nak(ctx)
	a.counter.WithLabelValues(a.Subject(), "nak").Inc()
	return err
}

func (a *ackableWrapper) NackWithDelay(ctx context.Context, delay time.Duration) error {
	err := a.Msg.(peanats.Ackable).NackWithDelay(ctx, delay)
	a.counter.WithLabelValues(a.Subject(), "nak").Inc()
	return err
}

func (a *ackableWrapper) Term(ctx context.Context) error {
	err := a.Msg.(peanats.Ackable).Term(ctx)
	a.counter.WithLabelValues(a.Subject(), "term").Inc()
	return err
}

func (a *ackableWrapper) TermWithReason(ctx context.Context, reason string) error {
	err := a.Msg.(peanats.Ackable).TermWithReason(ctx, reason)
	a.counter.WithLabelValues(a.Subject(), "term").Inc()
	return err
}

func (a *ackableWrapper) InProgress(ctx context.Context) error {
	err := a.Msg.(peanats.Ackable).InProgress(ctx)
	a.counter.WithLabelValues(a.Subject(), "in_progress").Inc()
	return err
}

// Metadata implements Metadatable interface if the underlying message supports it and panics
// otherwise. In practice, messages implementing Ackable always implement Metadatable as well.
func (a *ackableWrapper) Metadata() (*jetstream.MsgMetadata, error) {
	if m, ok := a.Msg.(peanats.Metadatable); ok {
		return m.Metadata()
	} else {
		panic("ackableWrapper: underlying message does not implement Metadatable interface")
	}
}
