# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Peanats is a generic typed handlers framework for NATS messaging in Go. It provides type-parametrized interfaces for various messaging patterns including publisher/subscriber, producer/consumer, and request/reply patterns.

## Branching

- Use branch naming format: `<github-handle>/<branch-name>`
- Example: `octocat/feature-name`, `octocat/bug-fix`

## Development Commands

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./publisher
go test ./subscriber
go test ./bucket

# Run a specific test
go test -run TestSpecificFunction ./path/to/package
```

### Mock Generation
```bash
# Generate mocks using mockery (configured in .mockery.yaml)
mockery
```

### Building
```bash
# Build all packages
go build ./...

# Verify module dependencies
go mod verify
go mod tidy
```

## Architecture Overview

### Core Components

- **Connection Management** (`nats.go`): Core interfaces for NATS connections with typed messaging support
- **Message System** (`msg.go`): Typed message interfaces supporting both regular NATS and JetStream messages
- **Codec System** (`codec.go`): Multi-format serialization (JSON, YAML, MessagePack, Protocol Buffers) with content-type based selection
- **Submitter Pattern** (`subm.go`): Async task execution abstraction for decoupled processing

### Messaging Patterns

Each package implements a specific messaging pattern with full type safety:

- **`/publisher/`**: Type-safe message publishing with automatic serialization based on content-type headers
- **`/subscriber/`**: Channel-based message consumption with configurable buffering and termination signals  
- **`/consumer/`**: JetStream pull consumer implementation for durable message processing
- **`/requester/`**: Request/reply pattern with support for single requests and streaming responses
- **`/bucket/`**: Typed key-value store wrapper around JetStream KeyValue with versioning and watching

### Integration Modules (`/contrib/`)

- **`/otel/`**: OpenTelemetry tracing and metrics integration
- **`/slog/`**: Structured logging with Go's slog package
- **`/pond/`**: Worker pool integration using Alitto Pond
- **`/raft/`**: Raft consensus algorithm integration
- **`/acknak/`**: Message acknowledgment helpers
- **`/muxer/`**: Message routing and multiplexing utilities

### Key Design Principles

1. **Type Safety**: Heavy use of Go generics (requires Go 1.24+) for compile-time type checking of message payloads
2. **Content-Type Aware**: Automatic codec selection based on NATS message headers
3. **Async Processing**: Built-in support for submitter patterns to decouple message handling
4. **Observability**: First-class OpenTelemetry and structured logging support
5. **Modularity**: Clear separation between core functionality and optional integrations

### Testing Structure

- Comprehensive test coverage with `*_test.go` files alongside each package
- Mock implementations generated using Mockery in `/internal/xmock/`
- Test utilities in `/internal/xtestutil/` for server setup and common test operations
- Examples in `/examples/` demonstrating pub/sub and client/server patterns

### Dependencies

- **nats.go**: Core NATS client library
- **JetStream**: For advanced messaging features (consumers, key-value store)
- **OpenTelemetry**: For observability and tracing
- **Various serialization libraries**: msgpack, protobuf, yaml for multi-format support

## Documentation

- Keep `CLAUDE.md` up to date during the work
- Update Documentation/Notes section of `CLAUDE.md` with any relevant technical notes
    - use 3rd level headings (###) for new sections
    - keep notes concise and focused on technical details
    - avoid lengthy explanations, focus on key points
    - use bullet points for clarity
    - keep notes human readable but also machine-readable for future AI interaction
- Update Documentation/Changelog section of `CLAUDE.md` with any significant changes
- Reference notes before starting new tasks for project context
- Keep notes in Markdown format for future reference
- Notes don't need to be fully human-readable, the main intent is to capture
  technical details, decisions, and context for future AI interaction

### Notes

#### Initial Architecture Analysis

- Generic typed handlers framework for NATS messaging in Go
- Heavy use of Go 1.24+ generics for compile-time type safety
- 5 core messaging patterns: publisher, subscriber, consumer, requester, bucket
- Multi-format codec system with content-type aware serialization
- Async processing via submitter patterns for decoupled handling
- Comprehensive contrib package ecosystem for integrations
- Production-ready with extensive test coverage and mock generation

#### Development Tooling

- Standard Go toolchain with go test, go build commands
- Mockery for mock generation (configured in .mockery.yaml)
- No custom build scripts or CI/CD detected
- Test structure follows Go conventions with comprehensive coverage
- Examples in /examples/ for pub/sub and client/server patterns

### Changelog

- 2025-05-26: Created initial CLAUDE.md with architecture overview and development commands
- 2025-05-26: Adopted note taking practice with Notes and Changelog sections
