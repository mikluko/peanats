package pearaft

import (
	"context"
	"fmt"
	"sync"

	"github.com/nats-io/graft"
	"github.com/nats-io/nats.go"
)

type State = graft.State

const (
	Follower  = graft.FOLLOWER
	Candidate = graft.CANDIDATE
	Leader    = graft.LEADER
	Closed    = graft.CLOSED
)

type Raft interface {
	State() State
	IsLeader() bool
	Wait(context.Context, State) error
}

func NewRaft(nc *nats.Conn, opts ...Option) (Raft, error) {
	p := params{
		name: "default",
		size: 3,
		ctx:  context.Background(),
		schh: func(from State, to State) error { return nil },
		errh: func(err error) { panic(err) },
		log:  "/dev/null",
	}
	for _, o := range opts {
		o(&p)
	}
	r := raftImpl{
		schh: p.schh,
		errh: p.errh,
		chc:  make(chan graft.StateChange),
		che:  make(chan error),
		cond: sync.Cond{L: new(sync.Mutex)},
	}
	info := graft.ClusterInfo{
		Name: p.name,
		Size: p.size,
	}
	hnd := graft.NewChanHandler(r.chc, r.che)
	rpc, err := graft.NewNatsRpcFromConn(nc)
	if err != nil {
		return nil, err
	}
	r.node, err = graft.New(info, hnd, rpc, p.log)
	if err != nil {
		return nil, err
	}
	go r.loop(p.ctx)
	return &r, nil
}

type params struct {
	size int
	name string
	ctx  context.Context
	schh func(from State, to State) error
	errh func(error)
	log  string
}

type Option func(*params)

func WithSize(size int) Option {
	return func(o *params) {
		o.size = size
	}
}

func WithName(name string) Option {
	return func(o *params) {
		o.name = name
	}
}

func WithContext(ctx context.Context) Option {
	return func(o *params) {
		o.ctx = ctx
	}
}

func WithErrorHandler(f func(error)) Option {
	return func(o *params) {
		o.errh = f
	}
}

func WithStateChangeHandler(f func(from State, to State) error) Option {
	return func(o *params) {
		o.schh = f
	}
}

func WithLog(log string) Option {
	return func(o *params) {
		o.log = log
	}
}

type raftImpl struct {
	node *graft.Node

	errh func(error)
	schh func(from State, to State) error

	chc  chan graft.StateChange
	che  chan error
	cond sync.Cond
}

func (r *raftImpl) State() State {
	if r.node == nil {
		return Closed
	}
	return r.node.State()
}

func (r *raftImpl) IsLeader() bool {
	return r.State() == Leader
}

func (r *raftImpl) Wait(ctx context.Context, state State) error {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()
	for r.State() != state {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			r.cond.Wait()
		}
	}
	return nil
}

func (r *raftImpl) loop(ctx context.Context) {
	for {
		select {
		case c := <-r.chc:
			err := r.schh(c.From, c.To)
			if err != nil {
				r.errh(fmt.Errorf("state change handler errored: %w", err))
			}
			r.cond.L.Lock()
			r.cond.Broadcast()
			r.cond.L.Unlock()
		case err := <-r.che:
			r.errh(err)
		case <-ctx.Done():
			r.node.Close()
			close(r.che)
			close(r.chc)
			r.cond.L.Lock()
			r.cond.Broadcast()
			r.cond.L.Unlock()
			return
		}
	}
}
