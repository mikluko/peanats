# Upgrading

## v0.21.x → v0.22.0

### Summary

The `codec/` and `transport/` packages have been extracted from the root `peanats` package to
resolve naming collisions and establish clearer dependency boundaries. This is a breaking change
with no deprecated aliases — all consumers must update imports and symbol names.

Additionally, a queue subscription bug was fixed: `Subscribe()` and `SubscribeHandler()` had
inverted conditions where `queue=""` incorrectly called queue methods and vice versa.

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
| `peanats.SubscribeHandlerSubmitter(subm)`    | `transport.SubscribeHandlerSubmitter(subm)`    |
| `peanats.SubscribeHandlerErrorHandler(errh)` | `transport.SubscribeHandlerErrorHandler(errh)` |
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
  -e 's/peanats\.SubscribeHandlerSubmitter/transport.SubscribeHandlerSubmitter/g' \
  -e 's/peanats\.SubscribeHandlerErrorHandler/transport.SubscribeHandlerErrorHandler/g' \
  -e 's/peanats\.SubscribeChanOption/transport.SubscribeChanOption/g' \
  -e 's/peanats\.SubscribeChanQueue/transport.SubscribeChanQueue/g' \
  {} +

# Fix imports (adds codec/ and transport/, removes unused peanats imports)
goimports -w .
```

> **Note:** The `sed` commands use specific-to-general ordering so that longer
> symbol names (e.g. `ContentTypeJson`) are matched before shorter ones
> (e.g. `ContentType`). Review the diff before committing.
