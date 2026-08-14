package trace

import (
	"strings"
	"unicode/utf8"

	"go.opentelemetry.io/otel/attribute"
)

// sanitizeString returns s unchanged when it is valid UTF-8, and otherwise a
// copy with each run of invalid bytes replaced by the Unicode replacement
// character. OTLP string fields must be valid UTF-8: one invalid value fails
// protobuf marshaling of the whole span batch on export.
func sanitizeString(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.ToValidUTF8(s, "�")
}

// sanitizeError returns err unchanged when its message is valid UTF-8, and
// otherwise an error whose message is the sanitized one and which unwraps to
// err, so span.RecordError never attaches an invalid exception.message
// attribute.
func sanitizeError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if utf8.ValidString(msg) {
		return err
	}
	return &sanitizedError{msg: strings.ToValidUTF8(msg, "�"), err: err}
}

type sanitizedError struct {
	msg string
	err error
}

func (e *sanitizedError) Error() string { return e.msg }

func (e *sanitizedError) Unwrap() error { return e.err }

// sanitizeAttrs returns attrs unchanged when every key and string value is
// valid UTF-8, and otherwise a copy with the offending strings sanitized.
// Non-string values pass through untouched.
func sanitizeAttrs(attrs []attribute.KeyValue) []attribute.KeyValue {
	for i := range attrs {
		if attrValid(attrs[i]) {
			continue
		}
		out := make([]attribute.KeyValue, len(attrs))
		copy(out, attrs[:i])
		for j := i; j < len(attrs); j++ {
			out[j] = sanitizeAttr(attrs[j])
		}
		return out
	}
	return attrs
}

func attrValid(kv attribute.KeyValue) bool {
	if !utf8.ValidString(string(kv.Key)) {
		return false
	}
	switch kv.Value.Type() {
	case attribute.STRING:
		return utf8.ValidString(kv.Value.AsString())
	case attribute.STRINGSLICE:
		for _, s := range kv.Value.AsStringSlice() {
			if !utf8.ValidString(s) {
				return false
			}
		}
	}
	return true
}

func sanitizeAttr(kv attribute.KeyValue) attribute.KeyValue {
	key := attribute.Key(sanitizeString(string(kv.Key)))
	switch kv.Value.Type() {
	case attribute.STRING:
		return key.String(sanitizeString(kv.Value.AsString()))
	case attribute.STRINGSLICE:
		values := kv.Value.AsStringSlice()
		for i := range values {
			values[i] = sanitizeString(values[i])
		}
		return key.StringSlice(values)
	default:
		return attribute.KeyValue{Key: key, Value: kv.Value}
	}
}
