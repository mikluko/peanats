package raft

import (
	"context"
	"sync"

	"github.com/nats-io/graft"
	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
)

type State = graft.State

const (
	RaftStateFollower  = graft.FOLLOWER
	RaftStateCandidate = graft.CANDIDATE
	RaftStateLeader    = graft.LEADER
	RaftStateClosed    = graft.CLOSED
)

type Raft interface {
	State() State
	IsLeader() bool
	Wait(context.Context, State) error
}

func New(nc *nats.Conn, opts ...Option) (Raft, error) {
	p := RaftParams{
		name: "default",
		size: 3,
		ctx:  context.Background(),
		schh: nullStateChangeHandlerImpl{},
		errh: peanats.DefaultErrorHandler,
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

type StateChangeHandler interface {
	HandleStateChange(ctx context.Context, from State, to State) error
}

type nullStateChangeHandlerImpl struct{}

func (n nullStateChangeHandlerImpl) HandleStateChange(_ context.Context, _ State, _ State) error {
	return nil
}

type RaftParams struct {
	size int
	name string
	ctx  context.Context
	schh StateChangeHandler
	errh peanats.ErrorHandler
	log  string
}

type Option func(*RaftParams)

func WithSize(size int) Option {
	return func(o *RaftParams) {
		o.size = size
	}
}

func WithName(name string) Option {
	return func(o *RaftParams) {
		o.name = name
	}
}

func WithContext(ctx context.Context) Option {
	return func(o *RaftParams) {
		o.ctx = ctx
	}
}

func WithErrorHandler(errh peanats.ErrorHandler) Option {
	return func(o *RaftParams) {
		o.errh = errh
	}
}

func WithStateChangeHandler(schh StateChangeHandler) Option {
	return func(o *RaftParams) {
		o.schh = schh
	}
}

func WithLog(log string) Option {
	return func(o *RaftParams) {
		o.log = log
	}
}

type raftImpl struct {
	node *graft.Node

	errh peanats.ErrorHandler
	schh StateChangeHandler

	chc  chan graft.StateChange
	che  chan error
	cond sync.Cond
}

func (r *raftImpl) State() State {
	if r.node == nil {
		return RaftStateClosed
	}
	return r.node.State()
}

func (r *raftImpl) IsLeader() bool {
	return r.State() == RaftStateLeader
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
			err := r.schh.HandleStateChange(ctx, c.From, c.To)
			if err != nil {
				r.errh.HandleError(ctx, err)
			}
			r.cond.L.Lock()
			r.cond.Broadcast()
			r.cond.L.Unlock()
		case err := <-r.che:
			r.errh.HandleError(ctx, err)
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
