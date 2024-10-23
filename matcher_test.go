package peanats

import "testing"

func TestMatcher(t *testing.T) {
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
			matcher := NewMatcher(tc.pattern)
			if matcher.Match(tc.subject) != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, !tc.expected)
			}
		})
	}
}
