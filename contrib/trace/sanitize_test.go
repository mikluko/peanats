package trace

import (
	"errors"
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestSanitizeString(t *testing.T) {
	t.Run("valid passes through", func(t *testing.T) {
		s := "check.dns.result π€"
		assert.Equal(t, s, sanitizeString(s))
	})
	t.Run("invalid run replaced", func(t *testing.T) {
		got := sanitizeString("a\xff\xfeb")
		assert.Equal(t, "a�b", got)
		assert.True(t, utf8.ValidString(got))
	})
	t.Run("truncated rune replaced", func(t *testing.T) {
		got := sanitizeString("π\xe2\x82")
		assert.Equal(t, "π�", got)
		assert.True(t, utf8.ValidString(got))
	})
}

func TestSanitizeError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.NoError(t, sanitizeError(nil))
	})
	t.Run("valid message passes through", func(t *testing.T) {
		err := errors.New("plain failure")
		assert.Same(t, err, sanitizeError(err))
	})
	t.Run("invalid message sanitized and unwraps to original", func(t *testing.T) {
		orig := fmt.Errorf("unexpected payload: %s", []byte{0xff, 0xfe})
		got := sanitizeError(orig)
		assert.Equal(t, "unexpected payload: �", got.Error())
		assert.True(t, utf8.ValidString(got.Error()))
		assert.ErrorIs(t, got, orig)
	})
}

func TestSanitizeAttrs(t *testing.T) {
	t.Run("clean slice returned as is", func(t *testing.T) {
		in := []attribute.KeyValue{
			attribute.String("k", "v"),
			attribute.Int("n", 42),
			attribute.StringSlice("ss", []string{"a", "b"}),
		}
		out := sanitizeAttrs(in)
		assert.Same(t, &in[0], &out[0])
	})
	t.Run("offending strings sanitized, input untouched", func(t *testing.T) {
		in := []attribute.KeyValue{
			attribute.String("good", "value"),
			attribute.String("bad\xffkey", "v"),
			attribute.String("s", "bad\xfevalue"),
			attribute.StringSlice("ss", []string{"ok", "bad\xff"}),
			attribute.Int("n", 42),
		}
		out := sanitizeAttrs(in)
		assert.Equal(t, attribute.String("good", "value"), out[0])
		assert.Equal(t, attribute.String("bad�key", "v"), out[1])
		assert.Equal(t, attribute.String("s", "bad�value"), out[2])
		assert.Equal(t, attribute.StringSlice("ss", []string{"ok", "bad�"}), out[3])
		assert.Equal(t, attribute.Int("n", 42), out[4])
		assert.Equal(t, attribute.String("bad\xffkey", "v"), in[1], "input must not be mutated")
		assert.Equal(t, attribute.StringSlice("ss", []string{"ok", "bad\xff"}), in[3], "input must not be mutated")
	})
}

// attrMap indexes attributes by key for assertions.
func attrMap(attrs []attribute.KeyValue) map[string]attribute.Value {
	m := make(map[string]attribute.Value, len(attrs))
	for _, kv := range attrs {
		m[string(kv.Key)] = kv.Value
	}
	return m
}

// assertExportedStringsValid asserts every string the span hands the OTLP
// exporter is valid UTF-8: the span name, the status description, and every
// attribute key and string value on the span and its events. Invalid UTF-8 in
// any of them fails protobuf marshaling of the whole span batch on export.
func assertExportedStringsValid(t *testing.T, span tracetest.SpanStub) {
	t.Helper()
	assert.True(t, utf8.ValidString(span.Name), "span name %q", span.Name)
	assert.True(t, utf8.ValidString(span.Status.Description), "status description %q", span.Status.Description)
	assertAttrsValid(t, span.Attributes)
	for _, ev := range span.Events {
		assert.True(t, utf8.ValidString(ev.Name), "event name %q", ev.Name)
		assertAttrsValid(t, ev.Attributes)
	}
}

func assertAttrsValid(t *testing.T, attrs []attribute.KeyValue) {
	t.Helper()
	for _, kv := range attrs {
		assert.True(t, utf8.ValidString(string(kv.Key)), "attribute key %q", kv.Key)
		switch kv.Value.Type() {
		case attribute.STRING:
			assert.True(t, utf8.ValidString(kv.Value.AsString()), "value of %q", kv.Key)
		case attribute.STRINGSLICE:
			for _, s := range kv.Value.AsStringSlice() {
				assert.True(t, utf8.ValidString(s), "slice value of %q", kv.Key)
			}
		}
	}
}
