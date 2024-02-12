package peanats

import (
	"context"
	"errors"
	"runtime"

	"github.com/alitto/pond"
	"github.com/nats-io/nats.go"
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

	pool *pond.WorkerPool
	subs []Unsubscriber
	ch   chan *nats.Msg
	done chan struct{}
}

func (s *Server) Start() error {
	if s.Conn == nil {
		return errors.New("nats connection is not set")
	}
	if s.Handler == nil {
		return errors.New("handler is not set")
	}
	if len(s.ListenSubjects) == 0 {
		return errors.New("no listen subjects")
	}
	if s.Concurrency == 0 {
		s.Concurrency = runtime.NumCPU()*2 - 1
	}
	if s.BaseContext == nil {
		s.BaseContext = context.Background()
	}

	s.done = make(chan struct{})
	s.ch = make(chan *nats.Msg)
	s.pool = pond.New(s.Concurrency+1, 0, pond.MinWorkers(s.Concurrency+1))
	s.pool.Submit(func() {
		for msg := range s.ch {
			s.pool.Submit(func() {
				s.handle(msg)
			})
		}
	})

	s.subs = make([]Unsubscriber, 0, len(s.ListenSubjects))
	for _, subj := range s.ListenSubjects {
		var (
			sub Unsubscriber
			err error
		)
		if s.QueueName == "" {
			sub, err = s.Conn.ChanSubscribe(subj, s.ch)
		} else {
			sub, err = s.Conn.ChanQueueSubscribe(subj, s.QueueName, s.ch)
		}
		if err != nil {
			for i := range s.subs {
				_ = s.subs[i].Unsubscribe()
			}
			return err
		}
		s.subs = append(s.subs, sub)
	}

	return nil
}

func (s *Server) Shutdown() {
	for i := range s.subs {
		_ = s.subs[i].Unsubscribe()
	}
	close(s.ch)
	s.pool.StopAndWait()
	close(s.done)
}

func (s *Server) Wait() {
	<-s.done
}

func (s *Server) handle(msg *nats.Msg) {
	rq := requestImpl{
		msg: msg,
	}
	rq.ctx, rq.done = context.WithCancel(s.BaseContext)
	err := s.Handler.Serve(&publisherImpl{PublisherMsg: s.Conn, subject: msg.Reply}, &rq)
	if err != nil {
		panic(err)
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
