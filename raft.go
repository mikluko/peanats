package peanats

import (
	"context"
	"sync"

	"github.com/nats-io/graft"
	"github.com/nats-io/nats.go"
)

type RaftState = graft.State

const (
	RaftStateFollower  = graft.FOLLOWER
	RaftStateCandidate = graft.CANDIDATE
	RaftStateLeader    = graft.LEADER
	RaftStateClosed    = graft.CLOSED
)

type Raft interface {
	State() RaftState
	IsLeader() bool
	Wait(context.Context, RaftState) error
}

func NewRaft(nc *nats.Conn, opts ...RaftOption) (Raft, error) {
	p := RaftParams{
		name: "default",
		size: 3,
		ctx:  context.Background(),
		schh: nullStateChangeHandlerImpl{},
		errh: PanicErrorHandler{},
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
	HandleStateChange(ctx context.Context, from RaftState, to RaftState) error
}

type nullStateChangeHandlerImpl struct{}

func (n nullStateChangeHandlerImpl) HandleStateChange(_ context.Context, _ RaftState, _ RaftState) error {
	return nil
}

type RaftParams struct {
	size int
	name string
	ctx  context.Context
	schh StateChangeHandler
	errh ErrorHandler
	log  string
}

type RaftOption func(*RaftParams)

func RaftSize(size int) RaftOption {
	return func(o *RaftParams) {
		o.size = size
	}
}

func RaftName(name string) RaftOption {
	return func(o *RaftParams) {
		o.name = name
	}
}

func RaftContext(ctx context.Context) RaftOption {
	return func(o *RaftParams) {
		o.ctx = ctx
	}
}

func RaftErrorHandler(errh ErrorHandler) RaftOption {
	return func(o *RaftParams) {
		o.errh = errh
	}
}

func RaftStateChangeHandler(schh StateChangeHandler) RaftOption {
	return func(o *RaftParams) {
		o.schh = schh
	}
}

func WithLog(log string) RaftOption {
	return func(o *RaftParams) {
		o.log = log
	}
}

type raftImpl struct {
	node *graft.Node

	errh ErrorHandler
	schh StateChangeHandler

	chc  chan graft.StateChange
	che  chan error
	cond sync.Cond
}

func (r *raftImpl) State() RaftState {
	if r.node == nil {
		return RaftStateClosed
	}
	return r.node.State()
}

func (r *raftImpl) IsLeader() bool {
	return r.State() == RaftStateLeader
}

func (r *raftImpl) Wait(ctx context.Context, state RaftState) error {
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
				err = r.errh.HandleError(ctx, err)
				if err != nil {
					panic(err)
				}
			}
			r.cond.L.Lock()
			r.cond.Broadcast()
			r.cond.L.Unlock()
		case err := <-r.che:
			err = r.errh.HandleError(ctx, err)
			if err != nil {
				panic(err)
			}
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
