// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package jetstreammock

import (
	"time"

	"github.com/nats-io/nats.go/jetstream"
	mock "github.com/stretchr/testify/mock"
)

// NewKeyValueEntry creates a new instance of KeyValueEntry. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewKeyValueEntry(t interface {
	mock.TestingT
	Cleanup(func())
}) *KeyValueEntry {
	mock := &KeyValueEntry{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// KeyValueEntry is an autogenerated mock type for the KeyValueEntry type
type KeyValueEntry struct {
	mock.Mock
}

type KeyValueEntry_Expecter struct {
	mock *mock.Mock
}

func (_m *KeyValueEntry) EXPECT() *KeyValueEntry_Expecter {
	return &KeyValueEntry_Expecter{mock: &_m.Mock}
}

// Bucket provides a mock function for the type KeyValueEntry
func (_mock *KeyValueEntry) Bucket() string {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Bucket")
	}

	var r0 string
	if returnFunc, ok := ret.Get(0).(func() string); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(string)
	}
	return r0
}

// KeyValueEntry_Bucket_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Bucket'
type KeyValueEntry_Bucket_Call struct {
	*mock.Call
}

// Bucket is a helper method to define mock.On call
func (_e *KeyValueEntry_Expecter) Bucket() *KeyValueEntry_Bucket_Call {
	return &KeyValueEntry_Bucket_Call{Call: _e.mock.On("Bucket")}
}

func (_c *KeyValueEntry_Bucket_Call) Run(run func()) *KeyValueEntry_Bucket_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *KeyValueEntry_Bucket_Call) Return(s string) *KeyValueEntry_Bucket_Call {
	_c.Call.Return(s)
	return _c
}

func (_c *KeyValueEntry_Bucket_Call) RunAndReturn(run func() string) *KeyValueEntry_Bucket_Call {
	_c.Call.Return(run)
	return _c
}

// Created provides a mock function for the type KeyValueEntry
func (_mock *KeyValueEntry) Created() time.Time {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Created")
	}

	var r0 time.Time
	if returnFunc, ok := ret.Get(0).(func() time.Time); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(time.Time)
	}
	return r0
}

// KeyValueEntry_Created_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Created'
type KeyValueEntry_Created_Call struct {
	*mock.Call
}

// Created is a helper method to define mock.On call
func (_e *KeyValueEntry_Expecter) Created() *KeyValueEntry_Created_Call {
	return &KeyValueEntry_Created_Call{Call: _e.mock.On("Created")}
}

func (_c *KeyValueEntry_Created_Call) Run(run func()) *KeyValueEntry_Created_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *KeyValueEntry_Created_Call) Return(time1 time.Time) *KeyValueEntry_Created_Call {
	_c.Call.Return(time1)
	return _c
}

func (_c *KeyValueEntry_Created_Call) RunAndReturn(run func() time.Time) *KeyValueEntry_Created_Call {
	_c.Call.Return(run)
	return _c
}

// Delta provides a mock function for the type KeyValueEntry
func (_mock *KeyValueEntry) Delta() uint64 {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Delta")
	}

	var r0 uint64
	if returnFunc, ok := ret.Get(0).(func() uint64); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(uint64)
	}
	return r0
}

// KeyValueEntry_Delta_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delta'
type KeyValueEntry_Delta_Call struct {
	*mock.Call
}

// Delta is a helper method to define mock.On call
func (_e *KeyValueEntry_Expecter) Delta() *KeyValueEntry_Delta_Call {
	return &KeyValueEntry_Delta_Call{Call: _e.mock.On("Delta")}
}

func (_c *KeyValueEntry_Delta_Call) Run(run func()) *KeyValueEntry_Delta_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *KeyValueEntry_Delta_Call) Return(v uint64) *KeyValueEntry_Delta_Call {
	_c.Call.Return(v)
	return _c
}

func (_c *KeyValueEntry_Delta_Call) RunAndReturn(run func() uint64) *KeyValueEntry_Delta_Call {
	_c.Call.Return(run)
	return _c
}

// Key provides a mock function for the type KeyValueEntry
func (_mock *KeyValueEntry) Key() string {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Key")
	}

	var r0 string
	if returnFunc, ok := ret.Get(0).(func() string); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(string)
	}
	return r0
}

// KeyValueEntry_Key_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Key'
type KeyValueEntry_Key_Call struct {
	*mock.Call
}

// Key is a helper method to define mock.On call
func (_e *KeyValueEntry_Expecter) Key() *KeyValueEntry_Key_Call {
	return &KeyValueEntry_Key_Call{Call: _e.mock.On("Key")}
}

func (_c *KeyValueEntry_Key_Call) Run(run func()) *KeyValueEntry_Key_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *KeyValueEntry_Key_Call) Return(s string) *KeyValueEntry_Key_Call {
	_c.Call.Return(s)
	return _c
}

func (_c *KeyValueEntry_Key_Call) RunAndReturn(run func() string) *KeyValueEntry_Key_Call {
	_c.Call.Return(run)
	return _c
}

// Operation provides a mock function for the type KeyValueEntry
func (_mock *KeyValueEntry) Operation() jetstream.KeyValueOp {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Operation")
	}

	var r0 jetstream.KeyValueOp
	if returnFunc, ok := ret.Get(0).(func() jetstream.KeyValueOp); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(jetstream.KeyValueOp)
	}
	return r0
}

// KeyValueEntry_Operation_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Operation'
type KeyValueEntry_Operation_Call struct {
	*mock.Call
}

// Operation is a helper method to define mock.On call
func (_e *KeyValueEntry_Expecter) Operation() *KeyValueEntry_Operation_Call {
	return &KeyValueEntry_Operation_Call{Call: _e.mock.On("Operation")}
}

func (_c *KeyValueEntry_Operation_Call) Run(run func()) *KeyValueEntry_Operation_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *KeyValueEntry_Operation_Call) Return(keyValueOp jetstream.KeyValueOp) *KeyValueEntry_Operation_Call {
	_c.Call.Return(keyValueOp)
	return _c
}

func (_c *KeyValueEntry_Operation_Call) RunAndReturn(run func() jetstream.KeyValueOp) *KeyValueEntry_Operation_Call {
	_c.Call.Return(run)
	return _c
}

// Revision provides a mock function for the type KeyValueEntry
func (_mock *KeyValueEntry) Revision() uint64 {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Revision")
	}

	var r0 uint64
	if returnFunc, ok := ret.Get(0).(func() uint64); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(uint64)
	}
	return r0
}

// KeyValueEntry_Revision_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Revision'
type KeyValueEntry_Revision_Call struct {
	*mock.Call
}

// Revision is a helper method to define mock.On call
func (_e *KeyValueEntry_Expecter) Revision() *KeyValueEntry_Revision_Call {
	return &KeyValueEntry_Revision_Call{Call: _e.mock.On("Revision")}
}

func (_c *KeyValueEntry_Revision_Call) Run(run func()) *KeyValueEntry_Revision_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *KeyValueEntry_Revision_Call) Return(v uint64) *KeyValueEntry_Revision_Call {
	_c.Call.Return(v)
	return _c
}

func (_c *KeyValueEntry_Revision_Call) RunAndReturn(run func() uint64) *KeyValueEntry_Revision_Call {
	_c.Call.Return(run)
	return _c
}

// Value provides a mock function for the type KeyValueEntry
func (_mock *KeyValueEntry) Value() []byte {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Value")
	}

	var r0 []byte
	if returnFunc, ok := ret.Get(0).(func() []byte); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}
	return r0
}

// KeyValueEntry_Value_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Value'
type KeyValueEntry_Value_Call struct {
	*mock.Call
}

// Value is a helper method to define mock.On call
func (_e *KeyValueEntry_Expecter) Value() *KeyValueEntry_Value_Call {
	return &KeyValueEntry_Value_Call{Call: _e.mock.On("Value")}
}

func (_c *KeyValueEntry_Value_Call) Run(run func()) *KeyValueEntry_Value_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *KeyValueEntry_Value_Call) Return(bytes []byte) *KeyValueEntry_Value_Call {
	_c.Call.Return(bytes)
	return _c
}

func (_c *KeyValueEntry_Value_Call) RunAndReturn(run func() []byte) *KeyValueEntry_Value_Call {
	_c.Call.Return(run)
	return _c
}
