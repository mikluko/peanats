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

Typed interfaces, middleware system, and transport adapter around NATS and JetStream:

```go
tc, _ := transport.Wrap(nats.Connect(nats.DefaultURL))
tc.Publish(ctx, peanats.NewMsg(...)) // Takes peanats.Msg interface
```

### Layer 3: Peanats Packages

High-level typed APIs with automatic serialization and content-type detection:

```go
pub := publisher.New(tc)
pub.Publish(ctx, "subject", MyStruct{}, publisher.WithContentType(codec.JSON))
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
- **`muxer/`** - Message routing and multiplexing utilities

## Quick Start

### Basic Publisher/Subscriber

```go
package main

import (
    "context"
    "github.com/nats-io/nats.go"
    "github.com/mikluko/peanats/publisher"
    "github.com/mikluko/peanats/transport"
)

type MyMessage struct {
    ID   string `json:"id"`
    Data string `json:"data"`
}

func main() {
    // Connect to NATS
    conn, _ := nats.Connect(nats.DefaultURL)
    tc, _ := transport.Wrap(conn)

    // Publisher
    pub := publisher.New(tc)
    pub.Publish(context.Background(), "events", MyMessage{
        ID: "123", Data: "hello world",
    })
}
```

### Request/Reply Pattern

```go
tc, _ := transport.Wrap(nats.Connect(nats.DefaultURL))

req := requester.New[MyRequest, MyResponse](tc)
resp, _ := req.Request(ctx, "service.endpoint", MyRequest{Query: "data"})

for r := range resp {
    if r.Err != nil {
        log.Printf("Error: %v", r.Err)
        continue
    }
    println("Response:", r.Payload.Result)
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

### Middleware Chain

Middleware wraps message handlers to add cross-cutting concerns. Use `ChainMsgMiddleware`
to compose them:

```go
h := peanats.ChainMsgMiddleware(
    peanats.MsgHandlerFromArgHandler[MyMessage](&myHandler{}),
    logging.AccessLogMiddleware(logging.SlogLogger(slog.Default(), slog.LevelInfo)),
    prom.Middleware(prom.MiddlewareNamespace("myapp")),
)

ch, _ := subscriber.SubscribeChan(ctx, h)
sub, _ := tc.SubscribeChan(ctx, "events.>", ch)
```

### OpenTelemetry Tracing

```go
// Add tracing to publisher
pub := trace.Publisher(publisher.New(tc))

// Add tracing to requester
req := trace.Requester(requester.New[MyReq, MyResp](tc))
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

## Versioning

This project follows [EffVer](https://jacobtomlinson.dev/effver/) (Effort Versioning) rather than strict SemVer, using the **MACRO.MESO.MICRO** scheme:

- **MACRO** bumps signal fundamental rethinks — currently frozen at 0
- **MESO** bumps introduce new features and may include breaking changes
- **MICRO** bumps are bug fixes and non-breaking improvements

Peanats is on the **0.x track** indefinitely. There are no plans for a 1.0 release — the API is stable enough for production use but reserves the right to evolve. When breaking changes occur at MESO boundaries, migration guidance is provided in [UPGRADING.md](UPGRADING.md).

## AI-Assisted Development

This project is developed with substantial AI assistance ([Claude Code](https://docs.anthropic.com/en/docs/claude-code)). Every design decision, interface contract, and architectural trade-off is human-directed and human-reviewed. AI accelerates implementation; it does not drive it.

Contributions that use AI tooling are welcome — under the same standard. If you can't explain why the code exists, it doesn't belong here.

## License

MIT License - see [LICENSE](LICENSE) file for details.