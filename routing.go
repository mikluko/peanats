package peanats

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type Route interface {
	Match(string) bool
	Handler
}

func NewRoute(handle Handler, patterns ...string) Route {
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
	Handler
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

// Muxer is a handler multiplexer
type Muxer interface {
	Handler
	Add(Route)
}

func NewMuxer(routes ...Route) Muxer {
	return &muxerImpl{
		routes: routes,
	}
}

type muxerImpl struct {
	routes []Route
}

func (r *muxerImpl) Handle(ctx context.Context, d Dispatcher, m Message) {
	for _, route := range r.routes {
		if route.Match(m.Subject()) {
			route.Handle(ctx, d, m)
			return
		}
	}
	d.Error(ctx, fmt.Errorf("no route for: %s", m.Subject()))
}

func (r *muxerImpl) Add(route Route) {
	r.routes = append(r.routes, route)
}
