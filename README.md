What K, # Peanats

A generic typed handlers framework for NATS messaging in Go.

## Features

- Type-safe messaging with Go generics
- Multi-format serialization (JSON, YAML, MessagePack, Protocol Buffers)
- Publisher/Subscriber, Producer/Consumer, Request/Reply patterns
- JetStream support for durable messaging and key-value storage
- OpenTelemetry integration

## Quick Start

```go
import "github.com/mikluko/peanats"

// Type-safe publisher
publisher := peanats.NewPublisher[MyMessage](conn)
publisher.Publish(ctx, "subject", MyMessage{Data: "hello"})

// Type-safe subscriber
subscriber := peanats.NewSubscriber[MyMessage](conn)
ch := subscriber.Subscribe(ctx, "subject")
for msg := range ch {
    // msg is typed as MyMessage
}
```

## Requirements

- Go 1.24 or later

## Installation

```bash
go get github.com/mikluko/peanats
```

## Examples

See [`examples/`](examples/) for complete working examples:
- [`pubsub/`](examples/pubsub/) - Publisher/Subscriber pattern
- [`clisrv/`](examples/clisrv/) - Client/Server request/reply pattern