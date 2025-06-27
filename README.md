# Peanats

[![Go Reference](https://pkg.go.dev/badge/github.com/mikluko/peanats.svg)](https://pkg.go.dev/github.com/mikluko/peanats)
[![Go Report Card](https://goreportcard.com/badge/github.com/mikluko/peanats)](https://goreportcard.com/report/github.com/mikluko/peanats)
[![GitHub License](https://img.shields.io/github/license/mikluko/peanats)](https://github.com/mikluko/peanats/blob/main/LICENSE)
[![GitHub Release](https://img.shields.io/github/v/tag/mikluko/peanats?label=release)](https://github.com/mikluko/peanats/tags)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mikluko/peanats)](https://github.com/mikluko/peanats/blob/main/go.mod)
[![GitHub Issues](https://img.shields.io/github/issues/mikluko/peanats)](https://github.com/mikluko/peanats/issues)
[![GitHub Stars](https://img.shields.io/github/stars/mikluko/peanats)](https://github.com/mikluko/peanats/stargazers)

A generic typed handlers framework for NATS messaging in Go.

## Features

- **Type-safe messaging** with Go 1.24+ generics
- **Multi-format serialization** (JSON, YAML, MessagePack, Protocol Buffers)
- **Multiple messaging patterns**: Publisher/Subscriber, Producer/Consumer, Request/Reply
- **JetStream support** for durable messaging and key-value storage
- **Comprehensive integrations** via contrib packages
- **Production-ready** with extensive test coverage and observability

## Architecture

Peanats provides a **3-layer architecture** for NATS messaging:

### Layer 1: NATS Native

Direct use of NATS and JetStream client libraries:

```go
conn, _ := nats.Connect(nats.DefaultURL)
conn.Publish("subject", []byte("data"))
```

### Layer 2: Peanats Core

Typed interfaces and middleware system around NATS and JetStream:

```go
pconn := peanats.NewConnection(conn)
pconn.Publish(ctx, peanats.NewMsg(...)) // Takes peanats.Msg interface
```

### Layer 3: Peanats Packages

High-level typed APIs with automatic serialization and content-type detection:

```go
pub := publisher.New(pconn)
pub.Publish(ctx, "subject", MyStruct{}, publisher.WithContentType("application/json"))
```

**When to use each layer:**

- **Layer 3** (packages): Most applications - type-safe, automatic serialization
- **Layer 2** (core): Performance-critical or when you need direct control over message handling  
- **Layer 1** (native): Direct NATS features not yet wrapped by peanats

## Core Packages

### Messaging Patterns

- **`publisher/`** - Type-safe message publishing with automatic serialization
- **`subscriber/`** - Channel-based message consumption with configurable buffering
- **`consumer/`** - JetStream pull consumer implementation for durable processing
- **`requester/`** - Request/reply pattern with support for streaming responses
- **`bucket/`** - Typed key-value store wrapper around JetStream KeyValue

### Integration Packages (`contrib/`)

- **`trace/`** - OpenTelemetry tracing and metrics integration
- **`prom/`** - Prometheus metrics middleware for message processing
- **`logging/`** - Structured logging with Go's slog package
- **`acknak/`** - Message acknowledgment helpers for JetStream
- **`pond/`** - Worker pool integration using Alitto Pond
- **`raft/`** - Raft consensus algorithm integration
- **`muxer/`** - Message routing and multiplexing utilities

## Quick Start

### Basic Publisher/Subscriber

```go
package main

import (
    "context"
    "github.com/nats-io/nats.go"
    "github.com/mikluko/peanats"
    "github.com/mikluko/peanats/publisher"
    "github.com/mikluko/peanats/subscriber"
)

type MyMessage struct {
    ID   string `json:"id"`
    Data string `json:"data"`
}

func main() {
    // Connect to NATS
    conn, _ := nats.Connect(nats.DefaultURL)
    pconn := peanats.NewConnection(conn)
    
    // Publisher
    pub := publisher.New(pconn)
    pub.Publish(context.Background(), "events", MyMessage{
        ID: "123", Data: "hello world",
    })
    
    // Subscriber
    sub := subscriber.New[MyMessage](pconn)
    ch := sub.Subscribe(context.Background(), "events")
    
    for msg := range ch {
        // msg.Payload is typed as MyMessage
        println("Received:", msg.Payload.Data)
        msg.Ack(context.Background()) // Acknowledge if using JetStream
    }
}
```

### Request/Reply Pattern

```go
// Server
req := requester.New[MyRequest, MyResponse](pconn)
responses := req.Request(ctx, "service.endpoint", MyRequest{Query: "data"})

for response := range responses {
    if response.Err != nil {
        log.Printf("Error: %v", response.Err)
        continue
    }
    // response.Payload is typed as MyResponse
    println("Response:", response.Payload.Result)
}
```

### Key-Value Store

```go
// JetStream KV bucket
kv := bucket.New[MyData](js, "my-bucket")

// Put value
entry, _ := kv.Put(ctx, "key1", MyData{Value: "test"})

// Get value  
entry, _ = kv.Get(ctx, "key1")
// entry.Value() is typed as MyData
```

## Middleware & Observability

### Prometheus Metrics

```go
import "github.com/mikluko/peanats/contrib/prom"

// Add Prometheus middleware
middleware := prom.Middleware(
    prom.MiddlewareNamespace("myapp"),
    prom.MiddlewareSubsystem("events"),
)

// Use with subscriber
sub := subscriber.New[MyMessage](pconn, 
    subscriber.WithMiddleware(middleware))
```

### OpenTelemetry Tracing

```go
import "github.com/mikluko/peanats/contrib/trace"

// Add tracing to publisher
pub := trace.Publisher(publisher.New(pconn))

// Add tracing to requester
req := trace.Requester(requester.New[MyReq, MyResp](pconn))
```

### Structured Logging

```go
import "github.com/mikluko/peanats/contrib/logging"

middleware := logging.Middleware(slog.Default())
sub := subscriber.New[MyMessage](pconn,
    subscriber.WithMiddleware(middleware))
```

## Requirements

- **Go 1.24+** (uses generics extensively)
- **NATS Server** 2.9+ (for JetStream features)

## Installation

```bash
go get github.com/mikluko/peanats@latest
```

## Examples

See [`examples/`](examples/) for complete working examples:

- [`pubsub/`](examples/pubsub/) - Publisher/Subscriber pattern
- [`clisrv/`](examples/clisrv/) - Client/Server request/reply pattern

## Content-Type Support

Peanats automatically selects serialization format based on message content-type headers:

- `application/json` - JSON (default)
- `application/yaml` - YAML  
- `application/msgpack` - MessagePack
- `application/protobuf` - Protocol Buffers

## License

MIT License - see [LICENSE](LICENSE) file for details.