# Peanats

A generic typed handlers framework for NATS messaging in Go.

## Features

- Type-safe messaging with Go generics
- Multi-format serialization (JSON, YAML, MessagePack, Protocol Buffers)
- Publisher/Subscriber, Producer/Consumer, Request/Reply patterns
- JetStream support for durable messaging and key-value storage
- OpenTelemetry integration

## Architecture

Peanats provides a **3-layer architecture** for NATS messaging:

### Layer 1: NATS Native

Direct use of NATS and JetStream client libraries:

```go
conn, _ := nats.Connect(nats.DefaultURL)
conn.Publish("subject", []byte("data"))
```

### Layer 2: Peanats Core

Thin interface wrappers around NATS and JetStream messaging:

```go
pconn := peanats.NewConnection(conn)
pconn.Publish(ctx, peanats.NewMsg(...)) // Takes peanats.Msg interface
```

### Layer 3: Peanats Packages

High-level typed APIs with automatic serialization:

```go
pub := publisher.New(pconn)
pub.Publish(ctx, "subject", MyStruct{}, opts...) // Takes any type
```

**When to use each layer:**

- **Layer 3** (packages): Most applications - type-safe, automatic serialization
- **Layer 2** (core): Performance-critical or when you need direct control over message handling
- **Layer 1** (native): Direct NATS features not yet wrapped by peanats

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