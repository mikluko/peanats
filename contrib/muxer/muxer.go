package muxer

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mikluko/peanats"
)

func Middleware(routes ...Route) peanats.MsgMiddleware {
	mux := NewMuxer(routes...)
	return func(next peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, m peanats.Msg) error {
			err := mux.HandleMsg(ctx, m)
			if errors.Is(err, ErrNotFound) {
				return next.HandleMsg(ctx, m)
			}
			return err
		})
	}
}

func NewMuxer(routes ...Route) peanats.MsgHandler {
	return &muxerImpl{routes}
}

type Muxer interface {
	peanats.MsgHandler
	Add(Route)
}

type muxerImpl struct {
	routes []Route
}

func (r *muxerImpl) HandleMsg(ctx context.Context, m peanats.Msg) error {
	for _, route := range r.routes {
		if route.Match(m.Subject()) {
			return route.HandleMsg(ctx, m)
		}
	}
	return fmt.Errorf("%w: %s", ErrNotFound, m.Subject())
}

func (r *muxerImpl) Add(route Route) {
	r.routes = append(r.routes, route)
}

type Route interface {
	Match(string) bool
	peanats.MsgHandler
}

func NewRoute(handle peanats.MsgHandler, patterns ...string) Route {
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
	peanats.MsgHandler
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

var (
	ErrNotFound     = fmt.Errorf("no route found")
	NotFoundHandler = peanats.MsgHandlerFunc(func(_ context.Context, _ peanats.Msg) error {
		return ErrNotFound
	})
)
