package peanats

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
)

var (
	ErrMuxNotFound = errors.New("not found")
)

type muxEntry struct {
	handler Handler
	subject string
}

// ServeMux is a handler multiplexer
type ServeMux struct {
	m    map[string]muxEntry
	once sync.Once
}

func (m *ServeMux) init() {
	m.m = make(map[string]muxEntry)
}

func (m *ServeMux) Serve(pub Publisher, req Request) error {
	m.once.Do(m.init)
	entry, found := m.m[req.Subject()]
	if !found {
		return &Error{
			Code:    http.StatusNotFound,
			Message: http.StatusText(http.StatusNotFound),
		}
	}
	return entry.handler.Serve(pub, req)
}

func (m *ServeMux) Handle(f Handler, subjects ...string) error {
	m.once.Do(m.init)
	for i := range subjects {
		if _, found := m.m[subjects[i]]; found {
			return fmt.Errorf("duplicate subject: %q", subjects[i])
		}
		m.m[subjects[i]] = muxEntry{
			handler: f,
			subject: subjects[i],
		}
	}
	return nil
}

func (m *ServeMux) HandleFunc(f func(Publisher, Request) error, subjects ...string) error {
	return m.Handle(HandlerFunc(f), subjects...)
}
