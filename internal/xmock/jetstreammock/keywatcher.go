// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package jetstreammock

import (
	"github.com/nats-io/nats.go/jetstream"
	mock "github.com/stretchr/testify/mock"
)

// NewKeyWatcher creates a new instance of KeyWatcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewKeyWatcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *KeyWatcher {
	mock := &KeyWatcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// KeyWatcher is an autogenerated mock type for the KeyWatcher type
type KeyWatcher struct {
	mock.Mock
}

type KeyWatcher_Expecter struct {
	mock *mock.Mock
}

func (_m *KeyWatcher) EXPECT() *KeyWatcher_Expecter {
	return &KeyWatcher_Expecter{mock: &_m.Mock}
}

// Stop provides a mock function for the type KeyWatcher
func (_mock *KeyWatcher) Stop() error {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Stop")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func() error); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// KeyWatcher_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type KeyWatcher_Stop_Call struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
func (_e *KeyWatcher_Expecter) Stop() *KeyWatcher_Stop_Call {
	return &KeyWatcher_Stop_Call{Call: _e.mock.On("Stop")}
}

func (_c *KeyWatcher_Stop_Call) Run(run func()) *KeyWatcher_Stop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *KeyWatcher_Stop_Call) Return(err error) *KeyWatcher_Stop_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *KeyWatcher_Stop_Call) RunAndReturn(run func() error) *KeyWatcher_Stop_Call {
	_c.Call.Return(run)
	return _c
}

// Updates provides a mock function for the type KeyWatcher
func (_mock *KeyWatcher) Updates() <-chan jetstream.KeyValueEntry {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Updates")
	}

	var r0 <-chan jetstream.KeyValueEntry
	if returnFunc, ok := ret.Get(0).(func() <-chan jetstream.KeyValueEntry); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan jetstream.KeyValueEntry)
		}
	}
	return r0
}

// KeyWatcher_Updates_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Updates'
type KeyWatcher_Updates_Call struct {
	*mock.Call
}

// Updates is a helper method to define mock.On call
func (_e *KeyWatcher_Expecter) Updates() *KeyWatcher_Updates_Call {
	return &KeyWatcher_Updates_Call{Call: _e.mock.On("Updates")}
}

func (_c *KeyWatcher_Updates_Call) Run(run func()) *KeyWatcher_Updates_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *KeyWatcher_Updates_Call) Return(keyValueEntryCh <-chan jetstream.KeyValueEntry) *KeyWatcher_Updates_Call {
	_c.Call.Return(keyValueEntryCh)
	return _c
}

func (_c *KeyWatcher_Updates_Call) RunAndReturn(run func() <-chan jetstream.KeyValueEntry) *KeyWatcher_Updates_Call {
	_c.Call.Return(run)
	return _c
}
