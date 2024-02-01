package peanats

import (
	"context"
	"errors"
	"runtime"
	"sync"

	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

type Request interface {
	Context() context.Context
	Subject() string
	Reply() string
	Header() nats.Header
	Data() []byte
}

type Handler interface {
	Serve(Publisher, Request) error
}

type HandlerFunc func(Publisher, Request) error

func (f HandlerFunc) Serve(pub Publisher, req Request) error {
	return f(pub, req)
}

type Server struct {
	Conn           Connection
	Handler        Handler
	BaseContext    context.Context
	Concurrency    int
	QueueName      string
	ListenSubjects []string

	ch       chan *nats.Msg
	mux      sync.Mutex
	grp      *errgroup.Group
	ctx      context.Context
	shutdown context.CancelFunc
}

func (s *Server) Start() error {
	if !s.mux.TryLock() {
		return errors.New("server is already running")
	}
	if s.Conn == nil {
		s.mux.Unlock()
		return errors.New("nats connection is not set")
	}
	if s.Handler == nil {
		s.mux.Unlock()
		return errors.New("handler is not set")
	}
	if s.Concurrency == 0 {
		s.Concurrency = runtime.NumCPU()*2 - 1
	}
	if s.BaseContext == nil {
		s.BaseContext = context.Background()
	}

	s.ctx, s.shutdown = context.WithCancel(s.BaseContext)
	s.grp, s.ctx = errgroup.WithContext(s.ctx)
	s.ch = make(chan *nats.Msg, nats.DefaultMaxChanLen)

	for i := 0; i < s.Concurrency; i++ {
		s.grp.Go(func() error {
			return s.handleMsgLoop()
		})
	}

	var subs []Subscription
	for _, subj := range s.ListenSubjects {
		var (
			sub Subscription
			err error
		)
		if s.QueueName == "" {
			sub, err = s.Conn.ChanSubscribe(subj, s.ch)
		} else {
			sub, err = s.Conn.ChanQueueSubscribe(subj, s.QueueName, s.ch)
		}
		if err != nil {
			for i := range subs {
				_ = subs[i].Unsubscribe()
			}
			return err
		}
		subs = append(subs, sub)
	}

	return nil
}

func (s *Server) Shutdown() error {
	if s.mux.TryLock() {
		s.mux.Unlock()
		return errors.New("server is not running")
	}
	s.grp.Go(func() error {
		defer close(s.ch)
		defer s.shutdown()
		return s.Conn.Drain()
	})
	return nil
}

func (s *Server) Wait() error {
	if s.mux.TryLock() {
		s.mux.Unlock()
		return errors.New("server is not running")
	}
	return s.grp.Wait()
}

func (s *Server) handleMsg(msg *nats.Msg) {
	rq := requestImpl{
		msg: msg,
	}
	rq.ctx, rq.done = context.WithCancel(s.BaseContext)
	err := s.Handler.Serve(&publisher{msgpub: s.Conn, msg: msg}, &rq)
	if err != nil {
		panic(err)
	}
}

func (s *Server) handleMsgLoop() error {
	for {
		select {
		case <-s.ctx.Done():
			if errors.Is(s.ctx.Err(), context.Canceled) {
				return nil
			}
			return s.ctx.Err()
		case msg := <-s.ch:
			s.handleMsg(msg)
		}
	}
}

type requestImpl struct {
	ctx  context.Context
	done context.CancelFunc
	msg  *nats.Msg
}

func (r *requestImpl) Context() context.Context {
	return r.ctx
}

func (r *requestImpl) WithContext(ctx context.Context) Request {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(requestImpl)
	*r2 = *r
	r2.ctx, r2.done = context.WithCancel(ctx)
	return r2
}

func (r *requestImpl) Subject() string {
	return r.msg.Subject
}

func (r *requestImpl) Reply() string {
	return r.msg.Reply
}

func (r *requestImpl) Header() nats.Header {
	return r.msg.Header
}

func (r *requestImpl) Data() []byte {
	return r.msg.Data
}
