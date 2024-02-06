package stream

import (
	"context"
	"github.com/mikluko/peanats"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
)

type publisherMock struct {
	mock.Mock
}

func (p *publisherMock) Subject() string {
	args := p.Called()
	return args.String(0)
}

func (p *publisherMock) WithSubject(subj string) peanats.Publisher {
	args := p.Called(subj)
	return args.Get(0).(peanats.Publisher)
}

func (p *publisherMock) Header() *nats.Header {
	args := p.Called()
	return args.Get(0).(*nats.Header)
}

func (p *publisherMock) Publish(data []byte) error {
	args := p.Called(data)
	return args.Error(0)
}

func (p *publisherMock) PublishMsg(msg *nats.Msg) error {
	args := p.Called(msg)
	return args.Error(0)
}

type requestMock struct {
	mock.Mock
}

func (r *requestMock) Context() context.Context {
	args := r.Mock.Called()
	return args.Get(0).(context.Context)
}

func (r *requestMock) Subject() string {
	args := r.Mock.Called()
	return args.Get(0).(string)
}

func (r *requestMock) Reply() string {
	args := r.Mock.Called()
	return args.Get(0).(string)
}

func (r *requestMock) Header() nats.Header {
	args := r.Mock.Called()
	return args.Get(0).(nats.Header)
}

func (r *requestMock) Data() []byte {
	args := r.Mock.Called()
	return args.Get(0).([]byte)
}
