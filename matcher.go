package peanats

import (
	"regexp"
	"strings"
)

type Matcher interface {
	Match(string) bool
}

var _ Matcher = (*matcherImpl)(nil)

type matcherImpl struct {
	exp *regexp.Regexp
}

func (p *matcherImpl) Match(s string) bool {
	return p.exp.MatchString(s)
}

func NewMatcher(pat string) Matcher {
	exp := regexp.QuoteMeta(pat)
	if strings.Contains(exp, "*") {
		exp = strings.ReplaceAll(exp, "\\*", "[^.]+")
	}
	if strings.HasSuffix(exp, ">") {
		exp = strings.TrimSuffix(exp, ">") + "(.+)"
	}
	return &matcherImpl{
		regexp.MustCompile("^" + exp + "$"),
	}
}
