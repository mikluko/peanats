package xmux

import (
	"context"
	"fmt"

	"github.com/mikluko/peanats/xmsg"
)

// Muxer is a handler multiplexer
type Muxer interface {
	xmsg.MsgHandler
	Add(Route)
}

func NewMuxer(routes ...Route) Muxer {
	return &muxerImpl{
		routes: routes,
	}
}

var ErrNotFound = fmt.Errorf("no route found")

type muxerImpl struct {
	routes []Route
}

func (r *muxerImpl) HandleMsg(ctx context.Context, m xmsg.Msg) error {
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
