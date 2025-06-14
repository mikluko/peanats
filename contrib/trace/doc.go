// Package trace provides OpenTelemetry trace propagation utilities for NATS messaging.
//
// This package includes:
//   - Publisher wrapper that injects trace context into NATS message headers
//   - Middleware that extracts trace context from incoming NATS messages
//   - Automatic span creation with relevant attributes for messaging operations
//
// Example usage:
//
//	// Create a trace-aware publisher
//	pub := trace.NewPublisher(
//	    publisher.New(conn),
//	    trace.WithTracer(tracer),
//	)
//
//	// Use trace middleware for message handling
//	middleware := trace.Middleware(
//	    trace.WithMiddlewareTracer(tracer),
//	    trace.WithSpanKind(trace.SpanKindConsumer),
//	)
package trace
