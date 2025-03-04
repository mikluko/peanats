package peanats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteImpl_Match(t *testing.T) {
	tt := []struct {
		pattern  string
		subject  string
		expected bool
	}{
		{"parson.had.a.dog", "parson.had.a.dog", true},
		{"parson.had.a.cat", "parson.had.a.dog", false},
		{"parson.had.a.>", "parson.had.a.dog", true},
		{"parson.*.a.*", "parson.had.a.dog", true},
		{"parson.*.a.>", "parson.had.a.dog", true},
		{"parson.>.a.>", "parson.had.a.dog", false},
	}
	for _, tc := range tt {
		t.Run(tc.pattern, func(t *testing.T) {
			r := NewRoute(nil, tc.pattern)
			if tc.expected {
				assert.True(t, r.Match(tc.subject), "route %q should match subject %q", tc.pattern, tc.subject)
			} else {
				assert.False(t, r.Match(tc.subject), "route %q should not match subject %q", tc.pattern, tc.subject)
			}
		})
	}
}
