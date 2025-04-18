// Code generated by mockery. DO NOT EDIT.

package bucketmock

import (
	bucket "github.com/mikluko/peanats/bucket"
	mock "github.com/stretchr/testify/mock"
)

// Watcher is an autogenerated mock type for the Watcher type
type Watcher[T any] struct {
	mock.Mock
}

type Watcher_Expecter[T any] struct {
	mock *mock.Mock
}

func (_m *Watcher[T]) EXPECT() *Watcher_Expecter[T] {
	return &Watcher_Expecter[T]{mock: &_m.Mock}
}

// Next provides a mock function with no fields
func (_m *Watcher[T]) Next() (bucket.Entry[T], error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Next")
	}

	var r0 bucket.Entry[T]
	var r1 error
	if rf, ok := ret.Get(0).(func() (bucket.Entry[T], error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() bucket.Entry[T]); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(bucket.Entry[T])
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Watcher_Next_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Next'
type Watcher_Next_Call[T any] struct {
	*mock.Call
}

// Next is a helper method to define mock.On call
func (_e *Watcher_Expecter[T]) Next() *Watcher_Next_Call[T] {
	return &Watcher_Next_Call[T]{Call: _e.mock.On("Next")}
}

func (_c *Watcher_Next_Call[T]) Run(run func()) *Watcher_Next_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Watcher_Next_Call[T]) Return(_a0 bucket.Entry[T], _a1 error) *Watcher_Next_Call[T] {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Watcher_Next_Call[T]) RunAndReturn(run func() (bucket.Entry[T], error)) *Watcher_Next_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Stop provides a mock function with no fields
func (_m *Watcher[T]) Stop() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Stop")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Watcher_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type Watcher_Stop_Call[T any] struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
func (_e *Watcher_Expecter[T]) Stop() *Watcher_Stop_Call[T] {
	return &Watcher_Stop_Call[T]{Call: _e.mock.On("Stop")}
}

func (_c *Watcher_Stop_Call[T]) Run(run func()) *Watcher_Stop_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Watcher_Stop_Call[T]) Return(_a0 error) *Watcher_Stop_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Watcher_Stop_Call[T]) RunAndReturn(run func() error) *Watcher_Stop_Call[T] {
	_c.Call.Return(run)
	return _c
}

// NewWatcher creates a new instance of Watcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWatcher[T any](t interface {
	mock.TestingT
	Cleanup(func())
}) *Watcher[T] {
	mock := &Watcher[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
