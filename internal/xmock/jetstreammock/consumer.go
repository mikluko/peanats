// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package jetstreammock

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
	mock "github.com/stretchr/testify/mock"
)

// NewConsumer creates a new instance of Consumer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewConsumer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Consumer {
	mock := &Consumer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// Consumer is an autogenerated mock type for the Consumer type
type Consumer struct {
	mock.Mock
}

type Consumer_Expecter struct {
	mock *mock.Mock
}

func (_m *Consumer) EXPECT() *Consumer_Expecter {
	return &Consumer_Expecter{mock: &_m.Mock}
}

// CachedInfo provides a mock function for the type Consumer
func (_mock *Consumer) CachedInfo() *jetstream.ConsumerInfo {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for CachedInfo")
	}

	var r0 *jetstream.ConsumerInfo
	if returnFunc, ok := ret.Get(0).(func() *jetstream.ConsumerInfo); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*jetstream.ConsumerInfo)
		}
	}
	return r0
}

// Consumer_CachedInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CachedInfo'
type Consumer_CachedInfo_Call struct {
	*mock.Call
}

// CachedInfo is a helper method to define mock.On call
func (_e *Consumer_Expecter) CachedInfo() *Consumer_CachedInfo_Call {
	return &Consumer_CachedInfo_Call{Call: _e.mock.On("CachedInfo")}
}

func (_c *Consumer_CachedInfo_Call) Run(run func()) *Consumer_CachedInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Consumer_CachedInfo_Call) Return(consumerInfo *jetstream.ConsumerInfo) *Consumer_CachedInfo_Call {
	_c.Call.Return(consumerInfo)
	return _c
}

func (_c *Consumer_CachedInfo_Call) RunAndReturn(run func() *jetstream.ConsumerInfo) *Consumer_CachedInfo_Call {
	_c.Call.Return(run)
	return _c
}

// Consume provides a mock function for the type Consumer
func (_mock *Consumer) Consume(handler jetstream.MessageHandler, opts ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error) {
	var tmpRet mock.Arguments
	if len(opts) > 0 {
		tmpRet = _mock.Called(handler, opts)
	} else {
		tmpRet = _mock.Called(handler)
	}
	ret := tmpRet

	if len(ret) == 0 {
		panic("no return value specified for Consume")
	}

	var r0 jetstream.ConsumeContext
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(jetstream.MessageHandler, ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error)); ok {
		return returnFunc(handler, opts...)
	}
	if returnFunc, ok := ret.Get(0).(func(jetstream.MessageHandler, ...jetstream.PullConsumeOpt) jetstream.ConsumeContext); ok {
		r0 = returnFunc(handler, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.ConsumeContext)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(jetstream.MessageHandler, ...jetstream.PullConsumeOpt) error); ok {
		r1 = returnFunc(handler, opts...)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Consumer_Consume_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Consume'
type Consumer_Consume_Call struct {
	*mock.Call
}

// Consume is a helper method to define mock.On call
//   - handler jetstream.MessageHandler
//   - opts ...jetstream.PullConsumeOpt
func (_e *Consumer_Expecter) Consume(handler interface{}, opts ...interface{}) *Consumer_Consume_Call {
	return &Consumer_Consume_Call{Call: _e.mock.On("Consume",
		append([]interface{}{handler}, opts...)...)}
}

func (_c *Consumer_Consume_Call) Run(run func(handler jetstream.MessageHandler, opts ...jetstream.PullConsumeOpt)) *Consumer_Consume_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 jetstream.MessageHandler
		if args[0] != nil {
			arg0 = args[0].(jetstream.MessageHandler)
		}
		var arg1 []jetstream.PullConsumeOpt
		var variadicArgs []jetstream.PullConsumeOpt
		if len(args) > 1 {
			variadicArgs = args[1].([]jetstream.PullConsumeOpt)
		}
		arg1 = variadicArgs
		run(
			arg0,
			arg1...,
		)
	})
	return _c
}

func (_c *Consumer_Consume_Call) Return(consumeContext jetstream.ConsumeContext, err error) *Consumer_Consume_Call {
	_c.Call.Return(consumeContext, err)
	return _c
}

func (_c *Consumer_Consume_Call) RunAndReturn(run func(handler jetstream.MessageHandler, opts ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error)) *Consumer_Consume_Call {
	_c.Call.Return(run)
	return _c
}

// Fetch provides a mock function for the type Consumer
func (_mock *Consumer) Fetch(batch int, opts ...jetstream.FetchOpt) (jetstream.MessageBatch, error) {
	var tmpRet mock.Arguments
	if len(opts) > 0 {
		tmpRet = _mock.Called(batch, opts)
	} else {
		tmpRet = _mock.Called(batch)
	}
	ret := tmpRet

	if len(ret) == 0 {
		panic("no return value specified for Fetch")
	}

	var r0 jetstream.MessageBatch
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(int, ...jetstream.FetchOpt) (jetstream.MessageBatch, error)); ok {
		return returnFunc(batch, opts...)
	}
	if returnFunc, ok := ret.Get(0).(func(int, ...jetstream.FetchOpt) jetstream.MessageBatch); ok {
		r0 = returnFunc(batch, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.MessageBatch)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(int, ...jetstream.FetchOpt) error); ok {
		r1 = returnFunc(batch, opts...)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Consumer_Fetch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Fetch'
type Consumer_Fetch_Call struct {
	*mock.Call
}

// Fetch is a helper method to define mock.On call
//   - batch int
//   - opts ...jetstream.FetchOpt
func (_e *Consumer_Expecter) Fetch(batch interface{}, opts ...interface{}) *Consumer_Fetch_Call {
	return &Consumer_Fetch_Call{Call: _e.mock.On("Fetch",
		append([]interface{}{batch}, opts...)...)}
}

func (_c *Consumer_Fetch_Call) Run(run func(batch int, opts ...jetstream.FetchOpt)) *Consumer_Fetch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 int
		if args[0] != nil {
			arg0 = args[0].(int)
		}
		var arg1 []jetstream.FetchOpt
		var variadicArgs []jetstream.FetchOpt
		if len(args) > 1 {
			variadicArgs = args[1].([]jetstream.FetchOpt)
		}
		arg1 = variadicArgs
		run(
			arg0,
			arg1...,
		)
	})
	return _c
}

func (_c *Consumer_Fetch_Call) Return(messageBatch jetstream.MessageBatch, err error) *Consumer_Fetch_Call {
	_c.Call.Return(messageBatch, err)
	return _c
}

func (_c *Consumer_Fetch_Call) RunAndReturn(run func(batch int, opts ...jetstream.FetchOpt) (jetstream.MessageBatch, error)) *Consumer_Fetch_Call {
	_c.Call.Return(run)
	return _c
}

// FetchBytes provides a mock function for the type Consumer
func (_mock *Consumer) FetchBytes(maxBytes int, opts ...jetstream.FetchOpt) (jetstream.MessageBatch, error) {
	var tmpRet mock.Arguments
	if len(opts) > 0 {
		tmpRet = _mock.Called(maxBytes, opts)
	} else {
		tmpRet = _mock.Called(maxBytes)
	}
	ret := tmpRet

	if len(ret) == 0 {
		panic("no return value specified for FetchBytes")
	}

	var r0 jetstream.MessageBatch
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(int, ...jetstream.FetchOpt) (jetstream.MessageBatch, error)); ok {
		return returnFunc(maxBytes, opts...)
	}
	if returnFunc, ok := ret.Get(0).(func(int, ...jetstream.FetchOpt) jetstream.MessageBatch); ok {
		r0 = returnFunc(maxBytes, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.MessageBatch)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(int, ...jetstream.FetchOpt) error); ok {
		r1 = returnFunc(maxBytes, opts...)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Consumer_FetchBytes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FetchBytes'
type Consumer_FetchBytes_Call struct {
	*mock.Call
}

// FetchBytes is a helper method to define mock.On call
//   - maxBytes int
//   - opts ...jetstream.FetchOpt
func (_e *Consumer_Expecter) FetchBytes(maxBytes interface{}, opts ...interface{}) *Consumer_FetchBytes_Call {
	return &Consumer_FetchBytes_Call{Call: _e.mock.On("FetchBytes",
		append([]interface{}{maxBytes}, opts...)...)}
}

func (_c *Consumer_FetchBytes_Call) Run(run func(maxBytes int, opts ...jetstream.FetchOpt)) *Consumer_FetchBytes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 int
		if args[0] != nil {
			arg0 = args[0].(int)
		}
		var arg1 []jetstream.FetchOpt
		var variadicArgs []jetstream.FetchOpt
		if len(args) > 1 {
			variadicArgs = args[1].([]jetstream.FetchOpt)
		}
		arg1 = variadicArgs
		run(
			arg0,
			arg1...,
		)
	})
	return _c
}

func (_c *Consumer_FetchBytes_Call) Return(messageBatch jetstream.MessageBatch, err error) *Consumer_FetchBytes_Call {
	_c.Call.Return(messageBatch, err)
	return _c
}

func (_c *Consumer_FetchBytes_Call) RunAndReturn(run func(maxBytes int, opts ...jetstream.FetchOpt) (jetstream.MessageBatch, error)) *Consumer_FetchBytes_Call {
	_c.Call.Return(run)
	return _c
}

// FetchNoWait provides a mock function for the type Consumer
func (_mock *Consumer) FetchNoWait(batch int) (jetstream.MessageBatch, error) {
	ret := _mock.Called(batch)

	if len(ret) == 0 {
		panic("no return value specified for FetchNoWait")
	}

	var r0 jetstream.MessageBatch
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(int) (jetstream.MessageBatch, error)); ok {
		return returnFunc(batch)
	}
	if returnFunc, ok := ret.Get(0).(func(int) jetstream.MessageBatch); ok {
		r0 = returnFunc(batch)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.MessageBatch)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(int) error); ok {
		r1 = returnFunc(batch)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Consumer_FetchNoWait_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FetchNoWait'
type Consumer_FetchNoWait_Call struct {
	*mock.Call
}

// FetchNoWait is a helper method to define mock.On call
//   - batch int
func (_e *Consumer_Expecter) FetchNoWait(batch interface{}) *Consumer_FetchNoWait_Call {
	return &Consumer_FetchNoWait_Call{Call: _e.mock.On("FetchNoWait", batch)}
}

func (_c *Consumer_FetchNoWait_Call) Run(run func(batch int)) *Consumer_FetchNoWait_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 int
		if args[0] != nil {
			arg0 = args[0].(int)
		}
		run(
			arg0,
		)
	})
	return _c
}

func (_c *Consumer_FetchNoWait_Call) Return(messageBatch jetstream.MessageBatch, err error) *Consumer_FetchNoWait_Call {
	_c.Call.Return(messageBatch, err)
	return _c
}

func (_c *Consumer_FetchNoWait_Call) RunAndReturn(run func(batch int) (jetstream.MessageBatch, error)) *Consumer_FetchNoWait_Call {
	_c.Call.Return(run)
	return _c
}

// Info provides a mock function for the type Consumer
func (_mock *Consumer) Info(context1 context.Context) (*jetstream.ConsumerInfo, error) {
	ret := _mock.Called(context1)

	if len(ret) == 0 {
		panic("no return value specified for Info")
	}

	var r0 *jetstream.ConsumerInfo
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context) (*jetstream.ConsumerInfo, error)); ok {
		return returnFunc(context1)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context) *jetstream.ConsumerInfo); ok {
		r0 = returnFunc(context1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*jetstream.ConsumerInfo)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = returnFunc(context1)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Consumer_Info_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Info'
type Consumer_Info_Call struct {
	*mock.Call
}

// Info is a helper method to define mock.On call
//   - context1 context.Context
func (_e *Consumer_Expecter) Info(context1 interface{}) *Consumer_Info_Call {
	return &Consumer_Info_Call{Call: _e.mock.On("Info", context1)}
}

func (_c *Consumer_Info_Call) Run(run func(context1 context.Context)) *Consumer_Info_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		run(
			arg0,
		)
	})
	return _c
}

func (_c *Consumer_Info_Call) Return(consumerInfo *jetstream.ConsumerInfo, err error) *Consumer_Info_Call {
	_c.Call.Return(consumerInfo, err)
	return _c
}

func (_c *Consumer_Info_Call) RunAndReturn(run func(context1 context.Context) (*jetstream.ConsumerInfo, error)) *Consumer_Info_Call {
	_c.Call.Return(run)
	return _c
}

// Messages provides a mock function for the type Consumer
func (_mock *Consumer) Messages(opts ...jetstream.PullMessagesOpt) (jetstream.MessagesContext, error) {
	var tmpRet mock.Arguments
	if len(opts) > 0 {
		tmpRet = _mock.Called(opts)
	} else {
		tmpRet = _mock.Called()
	}
	ret := tmpRet

	if len(ret) == 0 {
		panic("no return value specified for Messages")
	}

	var r0 jetstream.MessagesContext
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(...jetstream.PullMessagesOpt) (jetstream.MessagesContext, error)); ok {
		return returnFunc(opts...)
	}
	if returnFunc, ok := ret.Get(0).(func(...jetstream.PullMessagesOpt) jetstream.MessagesContext); ok {
		r0 = returnFunc(opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.MessagesContext)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(...jetstream.PullMessagesOpt) error); ok {
		r1 = returnFunc(opts...)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Consumer_Messages_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Messages'
type Consumer_Messages_Call struct {
	*mock.Call
}

// Messages is a helper method to define mock.On call
//   - opts ...jetstream.PullMessagesOpt
func (_e *Consumer_Expecter) Messages(opts ...interface{}) *Consumer_Messages_Call {
	return &Consumer_Messages_Call{Call: _e.mock.On("Messages",
		append([]interface{}{}, opts...)...)}
}

func (_c *Consumer_Messages_Call) Run(run func(opts ...jetstream.PullMessagesOpt)) *Consumer_Messages_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 []jetstream.PullMessagesOpt
		var variadicArgs []jetstream.PullMessagesOpt
		if len(args) > 0 {
			variadicArgs = args[0].([]jetstream.PullMessagesOpt)
		}
		arg0 = variadicArgs
		run(
			arg0...,
		)
	})
	return _c
}

func (_c *Consumer_Messages_Call) Return(messagesContext jetstream.MessagesContext, err error) *Consumer_Messages_Call {
	_c.Call.Return(messagesContext, err)
	return _c
}

func (_c *Consumer_Messages_Call) RunAndReturn(run func(opts ...jetstream.PullMessagesOpt) (jetstream.MessagesContext, error)) *Consumer_Messages_Call {
	_c.Call.Return(run)
	return _c
}

// Next provides a mock function for the type Consumer
func (_mock *Consumer) Next(opts ...jetstream.FetchOpt) (jetstream.Msg, error) {
	var tmpRet mock.Arguments
	if len(opts) > 0 {
		tmpRet = _mock.Called(opts)
	} else {
		tmpRet = _mock.Called()
	}
	ret := tmpRet

	if len(ret) == 0 {
		panic("no return value specified for Next")
	}

	var r0 jetstream.Msg
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(...jetstream.FetchOpt) (jetstream.Msg, error)); ok {
		return returnFunc(opts...)
	}
	if returnFunc, ok := ret.Get(0).(func(...jetstream.FetchOpt) jetstream.Msg); ok {
		r0 = returnFunc(opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.Msg)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(...jetstream.FetchOpt) error); ok {
		r1 = returnFunc(opts...)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Consumer_Next_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Next'
type Consumer_Next_Call struct {
	*mock.Call
}

// Next is a helper method to define mock.On call
//   - opts ...jetstream.FetchOpt
func (_e *Consumer_Expecter) Next(opts ...interface{}) *Consumer_Next_Call {
	return &Consumer_Next_Call{Call: _e.mock.On("Next",
		append([]interface{}{}, opts...)...)}
}

func (_c *Consumer_Next_Call) Run(run func(opts ...jetstream.FetchOpt)) *Consumer_Next_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 []jetstream.FetchOpt
		var variadicArgs []jetstream.FetchOpt
		if len(args) > 0 {
			variadicArgs = args[0].([]jetstream.FetchOpt)
		}
		arg0 = variadicArgs
		run(
			arg0...,
		)
	})
	return _c
}

func (_c *Consumer_Next_Call) Return(msg jetstream.Msg, err error) *Consumer_Next_Call {
	_c.Call.Return(msg, err)
	return _c
}

func (_c *Consumer_Next_Call) RunAndReturn(run func(opts ...jetstream.FetchOpt) (jetstream.Msg, error)) *Consumer_Next_Call {
	_c.Call.Return(run)
	return _c
}
