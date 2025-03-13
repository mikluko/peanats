package xmux

import (
	"regexp"
	"strings"

	"github.com/mikluko/peanats/xmsg"
)

type Route interface {
	Match(string) bool
	xmsg.MsgHandler
}

func NewRoute(handle xmsg.MsgHandler, patterns ...string) Route {
	exps := make([]regexp.Regexp, 0, len(patterns))
	for _, pat := range patterns {
		pat = regexp.QuoteMeta(pat)
		if strings.Contains(pat, "*") {
			pat = strings.ReplaceAll(pat, "\\*", "[^.]+")
		}
		if strings.HasSuffix(pat, ">") {
			pat = strings.TrimSuffix(pat, ">") + "(.+)"
		}
		exps = append(exps, *regexp.MustCompile("^" + pat + "$"))
	}
	return &routeImpl{handle, exps}
}

type routeImpl struct {
	xmsg.MsgHandler
	exps []regexp.Regexp
}

func (r *routeImpl) Match(subj string) bool {
	for i := range r.exps {
		if r.exps[i].MatchString(subj) {
			return true
		}
	}
	return false
}
