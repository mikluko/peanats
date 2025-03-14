// Code generated by mockery. DO NOT EDIT.

package peabucketmock

import (
	peabucket "github.com/mikluko/peanats/peabucket"
	mock "github.com/stretchr/testify/mock"
)

// BucketWatcher is an autogenerated mock type for the BucketWatcher type
type BucketWatcher[T any] struct {
	mock.Mock
}

type BucketWatcher_Expecter[T any] struct {
	mock *mock.Mock
}

func (_m *BucketWatcher[T]) EXPECT() *BucketWatcher_Expecter[T] {
	return &BucketWatcher_Expecter[T]{mock: &_m.Mock}
}

// Next provides a mock function with no fields
func (_m *BucketWatcher[T]) Next() (peabucket.BucketEntry[T], error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Next")
	}

	var r0 peabucket.BucketEntry[T]
	var r1 error
	if rf, ok := ret.Get(0).(func() (peabucket.BucketEntry[T], error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() peabucket.BucketEntry[T]); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(peabucket.BucketEntry[T])
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BucketWatcher_Next_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Next'
type BucketWatcher_Next_Call[T any] struct {
	*mock.Call
}

// Next is a helper method to define mock.On call
func (_e *BucketWatcher_Expecter[T]) Next() *BucketWatcher_Next_Call[T] {
	return &BucketWatcher_Next_Call[T]{Call: _e.mock.On("Next")}
}

func (_c *BucketWatcher_Next_Call[T]) Run(run func()) *BucketWatcher_Next_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BucketWatcher_Next_Call[T]) Return(_a0 peabucket.BucketEntry[T], _a1 error) *BucketWatcher_Next_Call[T] {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *BucketWatcher_Next_Call[T]) RunAndReturn(run func() (peabucket.BucketEntry[T], error)) *BucketWatcher_Next_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Stop provides a mock function with no fields
func (_m *BucketWatcher[T]) Stop() error {
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

// BucketWatcher_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type BucketWatcher_Stop_Call[T any] struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
func (_e *BucketWatcher_Expecter[T]) Stop() *BucketWatcher_Stop_Call[T] {
	return &BucketWatcher_Stop_Call[T]{Call: _e.mock.On("Stop")}
}

func (_c *BucketWatcher_Stop_Call[T]) Run(run func()) *BucketWatcher_Stop_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BucketWatcher_Stop_Call[T]) Return(_a0 error) *BucketWatcher_Stop_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BucketWatcher_Stop_Call[T]) RunAndReturn(run func() error) *BucketWatcher_Stop_Call[T] {
	_c.Call.Return(run)
	return _c
}

// NewBucketWatcher creates a new instance of BucketWatcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBucketWatcher[T any](t interface {
	mock.TestingT
	Cleanup(func())
}) *BucketWatcher[T] {
	mock := &BucketWatcher[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
