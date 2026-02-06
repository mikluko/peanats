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

#### Prometheus Middleware Metadatable Interface Fix

- Fixed prometheus middleware wrapper losing Metadatable interface from wrapped messages
- ackableWrapper now implements Metadata() method that delegates to underlying message
- Prevents type assertion failures in trace middleware when JetStream messages are wrapped
- Issue occurred because JetStream messages implement both Ackable and Metadatable
- Prometheus wrapper only preserved Ackable, causing downstream middleware to fail

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
- Always use mockery for generating mocks when testing interfaces
- Extend .mockery.yaml config as needed for new interfaces requiring mocks
- No custom build scripts or CI/CD detected
- Test structure follows Go conventions with comprehensive coverage
- Examples in /examples/ for pub/sub and client/server patterns

#### Naming Convention Concerns

- Multiple "Publisher" interfaces at different abstraction levels cause import conflicts
- peanats.Publisher (low-level) vs publisher.Publisher (high-level) naming collision
- Consumer vs Subscriber package distinction unclear from names alone
- consumer/ = JetStream pull consumers, subscriber/ = core NATS subscriptions
- Potential user confusion when choosing between packages and interfaces
- Consider renaming strategies for future releases to improve clarity

#### Tracing Requester Implementation

- Complete tracing wrapper for requester.Requester interface
- Supports both single Request calls and streaming ResponseReceiver operations  
- Automatic trace context injection into request headers via OpenTelemetry propagation
- Proper span lifecycle management for streaming operations
- Error handling with span status and error recording
- Configurable span names, attributes, and tracers
- Full test coverage with mocked dependencies

#### Middleware Ordering and Testing

- ChainMsgMiddleware function applies middlewares in forward iteration order (i := range mw)
- This creates reversed execution order where last middleware in slice executes first (outermost)
- Visual middleware list order: bottom middleware becomes outermost wrapper
- Comprehensive test coverage in msg_test.go verifies execution order
- Tests use mockery-generated mocks with .Maybe() for optional method calls
- Package naming pattern: peanats_test to avoid import cycles with internal mocks

#### Trace Package Test Fixes

- Fixed span attribute assertions to search by key instead of assuming order
- Resolved context type mismatch issues by using mock.Anything instead of specific context types
- Added .Once() to mock expectations to ensure proper call sequencing
- Made Header() calls optional with .Maybe() to handle conditional execution paths

#### Tracing Publisher Events (v0.21.0)

- Publisher tracing converted from spans to events — publishes add events to existing spans
- Eliminates redundant root spans (e.g. 0.04ms publish span parenting entire trace tree)
- Removed: PublisherWithTracer, PublisherWithSpanName, PublisherWithNewRoot, PublisherWithLinks
- Added: PublisherWithEventName (default "peanats.publish"), PublisherWithAttributes
- Header injection still happens regardless of span state for cross-process propagation
- Design principle: fire-and-forget operations → events; operations with duration → spans
- Middleware and requester keep spans (message handling and request/reply have meaningful duration)

#### Requester Header Management Fix

- Fixed critical bug in RequestHeader function that was replacing headers instead of merging
- RequestHeader now properly merges headers using Header.Add() to preserve existing headers
- Tracing requester no longer strips Content-Type and other existing headers
- Added comprehensive test coverage for header merging behavior
- Fixed tracing integration to properly preserve user-provided headers

### Changelog

- 2026-02-06: v0.21.0 — Replaced publisher trace spans with events (breaking: removed span-related options)
- 2025-07-02: Fixed prometheus middleware to preserve Metadatable interface when wrapping messages

- 2025-05-26: Created initial CLAUDE.md with architecture overview and development commands
- 2025-05-26: Adopted note taking practice with Notes and Changelog sections
- 2025-06-18: Added complete tracing requester implementation with comprehensive tests
- 2025-06-18: Added comprehensive middleware ordering tests to verify reversed execution order
- 2025-06-18: Fixed trace package test failures with proper mock expectations and context handling
- 2025-06-18: Fixed critical RequestHeader bug that stripped existing headers instead of merging
