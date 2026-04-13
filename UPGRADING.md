# Upgrading

## v0.24.x → v0.25.0

### Summary

`contrib/prom` now truncates the `subject` label to the first three
dot-separated tokens by default, bounding Prometheus metric cardinality on
JetStream consumers whose subjects contain dynamic tails (tenant IDs, entity
IDs, sequence numbers). Prior releases used the raw subject, which could
exhaust memory within minutes on high-fanout streams (#25).

This is a label-schema break: any dashboard, alert, or recording rule keyed
on the full subject will see a truncated value after the upgrade. Existing
series at the old cardinality remain in the TSDB until they age out — queries
that join across the cut-over point should use regex matching on the label.

### Action required

Callers who need the previous behavior must restore it explicitly:

```go
// Pass-through: use the raw subject as-is (pre-0.25 default).
// WARNING: susceptible to the unbounded-cardinality OOM described in #25.
_ = prom.Middleware(
    prom.MiddlewareSubjectMapper(nil),
)
```

Callers who want a different bound can pick their own:

```go
// Keep the first 5 tokens instead of the default 3.
_ = prom.Middleware(
    prom.MiddlewareSubjectMapper(prom.SubjectDepth(5)),
)

// Or collapse the subject dimension entirely.
_ = prom.Middleware(
    prom.MiddlewareSubjectMapper(prom.SubjectConstant("")),
)

// Or a custom mapping function.
_ = prom.Middleware(
    prom.MiddlewareSubjectMapper(func(s string) string {
        // ... e.g. strip UUIDs, apply per-stream rules, etc.
        return s
    }),
)
```

### Dashboards & alerts

Subject values in recorded metrics will change from e.g.
`orders.v1.created.tenant-42.entity-abc` to `orders.v1.created`. Update any
PromQL selectors that pin the full subject. Most dashboards aggregating with
`sum by (subject) (...)` will keep working but will show fewer, coarser
series.

---

## v0.22.x → v0.23.0

### Summary

The `codec/` and `transport/` packages have been extracted from the root `peanats` package to
resolve naming collisions and establish clearer dependency boundaries. This is a breaking change
with no deprecated aliases — all consumers must update imports and symbol names.

Additionally, a queue subscription bug was fixed: `Subscribe()` and `SubscribeHandler()` had
inverted conditions where `queue=""` incorrectly called queue methods and vice versa.

### Deprecated packages

- **`contrib/raft`** — Deprecated and removed. The package had no consumers outside its own tests and wrapped `nats-io/graft` with minimal added value. If you depend on it, pin to v0.21.x and consider using `nats-io/graft` directly.

### New imports

| Package     | Import path                            |
|-------------|----------------------------------------|
| `codec`     | `github.com/mikluko/peanats/codec`     |
| `transport` | `github.com/mikluko/peanats/transport` |

### Removed symbols

#### Connection & transport

| Old (peanats)                                | New (transport)                                |
|----------------------------------------------|------------------------------------------------|
| `peanats.Unsubscriber`                       | `transport.Unsubscriber`                       |
| `peanats.Subscription`                       | `transport.Subscription`                       |
| `peanats.Connection`                         | `transport.Conn`                               |
| `peanats.Publisher`                          | Removed (use `transport.Conn` directly)        |
| `peanats.Requester`                          | Removed (use `transport.Conn` directly)        |
| `peanats.Subscriber`                         | Removed (use `transport.Conn` directly)        |
| `peanats.Drainer`                            | Removed (use `transport.Conn` directly)        |
| `peanats.Closer`                             | Removed (use `transport.Conn` directly)        |
| `peanats.WrapConnection(conn, errs...)`      | `transport.Wrap(conn, errs...)`                |
| `peanats.NewConnection(conn)`                | `transport.New(conn)`                          |
| `peanats.SubscribeOption`                    | `transport.SubscribeOption`                    |
| `peanats.SubscribeQueue(name)`               | `transport.SubscribeQueue(name)`               |
| `peanats.SubscribeHandlerOption`             | `transport.SubscribeHandlerOption`             |
| `peanats.SubscribeHandlerSubmitter(subm)`    | `transport.SubscribeHandlerDispatcher(disp)`   |
| `peanats.SubscribeHandlerErrorHandler(errh)` | `transport.SubscribeHandlerDispatcher(disp)`   |
| `peanats.SubscribeChanOption`                | `transport.SubscribeChanOption`                |
| `peanats.SubscribeChanQueue(name)`           | `transport.SubscribeChanQueue(name)`           |

#### Codec types & functions

| Old (peanats)                             | New (codec)                          |
|-------------------------------------------|--------------------------------------|
| `peanats.Codec`                           | `codec.Codec`                        |
| `peanats.ContentType`                     | `codec.ContentType`                  |
| `peanats.HeaderContentType`               | `codec.HeaderContentType`            |
| `peanats.DefaultContentType`              | `codec.Default`                      |
| `peanats.CodecContentType(c)`             | `codec.ForContentType(c)`            |
| `peanats.CodecHeader(h)`                  | `codec.ForHeader(h)`                 |
| `peanats.ContentTypeHeader(h)`            | `codec.TypeFromHeader(h)`            |
| `peanats.ContentTypeHeaderCopy(dst, src)` | `codec.TypeFromHeaderCopy(dst, src)` |
| `peanats.MarshalHeader(x, h)`             | `codec.MarshalHeader(x, h)`          |
| `peanats.UnmarshalHeader(data, x, h)`     | `codec.UnmarshalHeader(data, x, h)`  |

#### Content type constants

| Old (peanats)                  | New (codec)       |
|--------------------------------|-------------------|
| `peanats.ContentTypeJson`      | `codec.JSON`      |
| `peanats.ContentTypeYaml`      | `codec.YAML`      |
| `peanats.ContentTypeMsgpack`   | `codec.Msgpack`   |
| `peanats.ContentTypeProtojson` | `codec.Protojson` |
| `peanats.ContentTypePrototext` | `codec.Prototext` |
| `peanats.ContentTypeProtobin`  | `codec.Protobin`  |

### Constructor signature changes

| Function                               | Old parameter         | New parameter                                            |
|----------------------------------------|-----------------------|----------------------------------------------------------|
| `publisher.New(pub)`                   | `peanats.Publisher`   | `publisher.RawPublisher` (satisfied by `transport.Conn`) |
| `requester.New[RQ, RS](nc)`            | `peanats.Connection`  | `transport.Conn`                                         |
| `misc.DefaultContentTypeMiddleware(c)` | `peanats.ContentType` | `codec.ContentType`                                      |

### Before / After

**Before** (v0.21.x):

```go
import (
"github.com/nats-io/nats.go"

"github.com/mikluko/peanats"
"github.com/mikluko/peanats/publisher"
)

func main() {
nc, err := peanats.WrapConnection(nats.Connect(nats.DefaultURL))
if err != nil {
panic(err)
}
defer nc.Close()

pub := publisher.New(nc)
_ = pub.Publish(ctx, "subject", payload,
publisher.WithContentType(peanats.ContentTypeJson),
publisher.WithHeader(peanats.Header{
peanats.HeaderContentType: []string{"application/json"},
}),
)
}
```

**After** (v0.22.0):

```go
import (
"github.com/nats-io/nats.go"

"github.com/mikluko/peanats"
"github.com/mikluko/peanats/codec"
"github.com/mikluko/peanats/publisher"
"github.com/mikluko/peanats/transport"
)

func main() {
nc, err := transport.Wrap(nats.Connect(nats.DefaultURL))
if err != nil {
panic(err)
}
defer nc.Close()

pub := publisher.New(nc)
_ = pub.Publish(ctx, "subject", payload,
publisher.WithContentType(codec.JSON),
publisher.WithHeader(peanats.Header{
codec.HeaderContentType: []string{"application/json"},
}),
)
}
```

### Dispatcher (replaces Submitter + ErrorHandler)

`Submitter` and `ErrorHandler` have been replaced by a single `Dispatcher` interface:

```go
type Dispatcher interface {
    Dispatch(func() error)
    Wait(context.Context) error
}
```

| Old | New |
|-----|-----|
| `peanats.Submitter` | `peanats.Dispatcher` |
| `peanats.ErrorHandler` | `peanats.Dispatcher` |
| `peanats.SubmitterFunc` | Removed |
| `peanats.ErrorHandlerFunc` | Removed |
| `peanats.DefaultSubmitter` | `peanats.DefaultDispatcher` |
| `peanats.DefaultErrorHandler` | `peanats.DefaultDispatcher` |
| `subscriber.SubscribeSubmitter(subm)` | `subscriber.SubscribeDispatcher(disp)` |
| `subscriber.SubscribeErrorHandler(errh)` | Removed (use `SubscribeDispatcher`) |
| `consumer.ConsumeSubmitter(subm)` | `consumer.ConsumeDispatcher(disp)` |
| `consumer.ConsumeErrorHandler(errh)` | Removed (use `ConsumeDispatcher`) |
| `bucket.WatchSubmitter(subm)` | `bucket.WatchDispatcher(disp)` |
| `bucket.WatchErrorHandler(errh)` | Removed (use `WatchDispatcher`) |
| `transport.SubscribeHandlerSubmitter(subm)` | `transport.SubscribeHandlerDispatcher(disp)` |
| `transport.SubscribeHandlerErrorHandler(errh)` | Removed (use `SubscribeHandlerDispatcher`) |
| `pond.Submitter(n)` | `pond.Dispatcher(n)` |
| `pond.SubmitterPool(pool)` | `pond.DispatcherPool(pool)` |

**Behavior change:** `DefaultDispatcher` remains fail-fast — it logs via `slog.Error` and
panics on task errors, ensuring failures are never silent. For graceful error collection,
use `peanats.NewDispatcher()` or `pond.Dispatcher()` — these collect errors internally
and return them from `Wait()`. Call `Wait()` during shutdown to drain in-flight tasks.

### Quick migration

The bulk of the migration is mechanical renaming. The following `sed` + `goimports` pipeline handles
most cases:

```bash
# Run from the repository root of your project
find . -name '*.go' -exec sed -i'' \
  -e 's/peanats\.Subscription/transport.Subscription/g' \
  -e 's/peanats\.Unsubscriber/transport.Unsubscriber/g' \
  -e 's/peanats\.WrapConnection/transport.Wrap/g' \
  -e 's/peanats\.NewConnection/transport.New/g' \
  -e 's/peanats\.Connection/transport.Conn/g' \
  -e 's/peanats\.CodecContentType/codec.ForContentType/g' \
  -e 's/peanats\.CodecHeader/codec.ForHeader/g' \
  -e 's/peanats\.ContentTypeHeader\b/codec.TypeFromHeader/g' \
  -e 's/peanats\.ContentTypeHeaderCopy/codec.TypeFromHeaderCopy/g' \
  -e 's/peanats\.MarshalHeader/codec.MarshalHeader/g' \
  -e 's/peanats\.UnmarshalHeader/codec.UnmarshalHeader/g' \
  -e 's/peanats\.HeaderContentType/codec.HeaderContentType/g' \
  -e 's/peanats\.DefaultContentType/codec.Default/g' \
  -e 's/peanats\.ContentTypeJson/codec.JSON/g' \
  -e 's/peanats\.ContentTypeYaml/codec.YAML/g' \
  -e 's/peanats\.ContentTypeMsgpack/codec.Msgpack/g' \
  -e 's/peanats\.ContentTypeProtojson/codec.Protojson/g' \
  -e 's/peanats\.ContentTypePrototext/codec.Prototext/g' \
  -e 's/peanats\.ContentTypeProtobin/codec.Protobin/g' \
  -e 's/peanats\.ContentType/codec.ContentType/g' \
  -e 's/peanats\.Codec/codec.Codec/g' \
  -e 's/peanats\.SubscribeOption/transport.SubscribeOption/g' \
  -e 's/peanats\.SubscribeQueue/transport.SubscribeQueue/g' \
  -e 's/peanats\.SubscribeHandlerOption/transport.SubscribeHandlerOption/g' \
  -e 's/peanats\.SubscribeHandlerSubmitter/transport.SubscribeHandlerDispatcher/g' \
  -e 's/peanats\.SubscribeHandlerErrorHandler/transport.SubscribeHandlerDispatcher/g' \
  -e 's/peanats\.SubscribeChanOption/transport.SubscribeChanOption/g' \
  -e 's/peanats\.SubscribeChanQueue/transport.SubscribeChanQueue/g' \
  -e 's/peanats\.Submitter/peanats.Dispatcher/g' \
  -e 's/peanats\.ErrorHandler/peanats.Dispatcher/g' \
  -e 's/peanats\.DefaultSubmitter/peanats.DefaultDispatcher/g' \
  -e 's/peanats\.DefaultErrorHandler/peanats.DefaultDispatcher/g' \
  -e 's/subscriber\.SubscribeSubmitter/subscriber.SubscribeDispatcher/g' \
  -e 's/subscriber\.SubscribeErrorHandler/subscriber.SubscribeDispatcher/g' \
  -e 's/consumer\.ConsumeSubmitter/consumer.ConsumeDispatcher/g' \
  -e 's/consumer\.ConsumeErrorHandler/consumer.ConsumeDispatcher/g' \
  -e 's/bucket\.WatchSubmitter/bucket.WatchDispatcher/g' \
  -e 's/bucket\.WatchErrorHandler/bucket.WatchDispatcher/g' \
  -e 's/pond\.Submitter/pond.Dispatcher/g' \
  -e 's/pond\.SubmitterPool/pond.DispatcherPool/g' \
  {} +

# Fix imports (adds codec/ and transport/, removes unused peanats imports)
goimports -w .
```

> **Note:** The `sed` commands use specific-to-general ordering so that longer
> symbol names (e.g. `ContentTypeJson`) are matched before shorter ones
> (e.g. `ContentType`). Review the diff before committing.
