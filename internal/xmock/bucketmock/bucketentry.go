// Code generated by mockery. DO NOT EDIT.

package bucketmock

import (
	jetstream "github.com/nats-io/nats.go/jetstream"
	mock "github.com/stretchr/testify/mock"

	peanats "github.com/mikluko/peanats"

	time "time"
)

// BucketEntry is an autogenerated mock type for the BucketEntry type
type BucketEntry[T any] struct {
	mock.Mock
}

type BucketEntry_Expecter[T any] struct {
	mock *mock.Mock
}

func (_m *BucketEntry[T]) EXPECT() *BucketEntry_Expecter[T] {
	return &BucketEntry_Expecter[T]{mock: &_m.Mock}
}

// Bucket provides a mock function with no fields
func (_m *BucketEntry[T]) Bucket() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Bucket")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// BucketEntry_Bucket_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Bucket'
type BucketEntry_Bucket_Call[T any] struct {
	*mock.Call
}

// Bucket is a helper method to define mock.On call
func (_e *BucketEntry_Expecter[T]) Bucket() *BucketEntry_Bucket_Call[T] {
	return &BucketEntry_Bucket_Call[T]{Call: _e.mock.On("Bucket")}
}

func (_c *BucketEntry_Bucket_Call[T]) Run(run func()) *BucketEntry_Bucket_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BucketEntry_Bucket_Call[T]) Return(_a0 string) *BucketEntry_Bucket_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BucketEntry_Bucket_Call[T]) RunAndReturn(run func() string) *BucketEntry_Bucket_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Created provides a mock function with no fields
func (_m *BucketEntry[T]) Created() time.Time {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Created")
	}

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// BucketEntry_Created_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Created'
type BucketEntry_Created_Call[T any] struct {
	*mock.Call
}

// Created is a helper method to define mock.On call
func (_e *BucketEntry_Expecter[T]) Created() *BucketEntry_Created_Call[T] {
	return &BucketEntry_Created_Call[T]{Call: _e.mock.On("Created")}
}

func (_c *BucketEntry_Created_Call[T]) Run(run func()) *BucketEntry_Created_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BucketEntry_Created_Call[T]) Return(_a0 time.Time) *BucketEntry_Created_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BucketEntry_Created_Call[T]) RunAndReturn(run func() time.Time) *BucketEntry_Created_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Delta provides a mock function with no fields
func (_m *BucketEntry[T]) Delta() uint64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Delta")
	}

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// BucketEntry_Delta_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delta'
type BucketEntry_Delta_Call[T any] struct {
	*mock.Call
}

// Delta is a helper method to define mock.On call
func (_e *BucketEntry_Expecter[T]) Delta() *BucketEntry_Delta_Call[T] {
	return &BucketEntry_Delta_Call[T]{Call: _e.mock.On("Delta")}
}

func (_c *BucketEntry_Delta_Call[T]) Run(run func()) *BucketEntry_Delta_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BucketEntry_Delta_Call[T]) Return(_a0 uint64) *BucketEntry_Delta_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BucketEntry_Delta_Call[T]) RunAndReturn(run func() uint64) *BucketEntry_Delta_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Header provides a mock function with no fields
func (_m *BucketEntry[T]) Header() peanats.Header {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Header")
	}

	var r0 peanats.Header
	if rf, ok := ret.Get(0).(func() peanats.Header); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(peanats.Header)
		}
	}

	return r0
}

// BucketEntry_Header_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Header'
type BucketEntry_Header_Call[T any] struct {
	*mock.Call
}

// Header is a helper method to define mock.On call
func (_e *BucketEntry_Expecter[T]) Header() *BucketEntry_Header_Call[T] {
	return &BucketEntry_Header_Call[T]{Call: _e.mock.On("Header")}
}

func (_c *BucketEntry_Header_Call[T]) Run(run func()) *BucketEntry_Header_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BucketEntry_Header_Call[T]) Return(_a0 peanats.Header) *BucketEntry_Header_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BucketEntry_Header_Call[T]) RunAndReturn(run func() peanats.Header) *BucketEntry_Header_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Key provides a mock function with no fields
func (_m *BucketEntry[T]) Key() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Key")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// BucketEntry_Key_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Key'
type BucketEntry_Key_Call[T any] struct {
	*mock.Call
}

// Key is a helper method to define mock.On call
func (_e *BucketEntry_Expecter[T]) Key() *BucketEntry_Key_Call[T] {
	return &BucketEntry_Key_Call[T]{Call: _e.mock.On("Key")}
}

func (_c *BucketEntry_Key_Call[T]) Run(run func()) *BucketEntry_Key_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BucketEntry_Key_Call[T]) Return(_a0 string) *BucketEntry_Key_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BucketEntry_Key_Call[T]) RunAndReturn(run func() string) *BucketEntry_Key_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Operation provides a mock function with no fields
func (_m *BucketEntry[T]) Operation() jetstream.KeyValueOp {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Operation")
	}

	var r0 jetstream.KeyValueOp
	if rf, ok := ret.Get(0).(func() jetstream.KeyValueOp); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(jetstream.KeyValueOp)
	}

	return r0
}

// BucketEntry_Operation_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Operation'
type BucketEntry_Operation_Call[T any] struct {
	*mock.Call
}

// Operation is a helper method to define mock.On call
func (_e *BucketEntry_Expecter[T]) Operation() *BucketEntry_Operation_Call[T] {
	return &BucketEntry_Operation_Call[T]{Call: _e.mock.On("Operation")}
}

func (_c *BucketEntry_Operation_Call[T]) Run(run func()) *BucketEntry_Operation_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BucketEntry_Operation_Call[T]) Return(_a0 jetstream.KeyValueOp) *BucketEntry_Operation_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BucketEntry_Operation_Call[T]) RunAndReturn(run func() jetstream.KeyValueOp) *BucketEntry_Operation_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Revision provides a mock function with no fields
func (_m *BucketEntry[T]) Revision() uint64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Revision")
	}

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// BucketEntry_Revision_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Revision'
type BucketEntry_Revision_Call[T any] struct {
	*mock.Call
}

// Revision is a helper method to define mock.On call
func (_e *BucketEntry_Expecter[T]) Revision() *BucketEntry_Revision_Call[T] {
	return &BucketEntry_Revision_Call[T]{Call: _e.mock.On("Revision")}
}

func (_c *BucketEntry_Revision_Call[T]) Run(run func()) *BucketEntry_Revision_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BucketEntry_Revision_Call[T]) Return(_a0 uint64) *BucketEntry_Revision_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BucketEntry_Revision_Call[T]) RunAndReturn(run func() uint64) *BucketEntry_Revision_Call[T] {
	_c.Call.Return(run)
	return _c
}

// Value provides a mock function with no fields
func (_m *BucketEntry[T]) Value() *T {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Value")
	}

	var r0 *T
	if rf, ok := ret.Get(0).(func() *T); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*T)
		}
	}

	return r0
}

// BucketEntry_Value_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Value'
type BucketEntry_Value_Call[T any] struct {
	*mock.Call
}

// Value is a helper method to define mock.On call
func (_e *BucketEntry_Expecter[T]) Value() *BucketEntry_Value_Call[T] {
	return &BucketEntry_Value_Call[T]{Call: _e.mock.On("Value")}
}

func (_c *BucketEntry_Value_Call[T]) Run(run func()) *BucketEntry_Value_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *BucketEntry_Value_Call[T]) Return(_a0 *T) *BucketEntry_Value_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *BucketEntry_Value_Call[T]) RunAndReturn(run func() *T) *BucketEntry_Value_Call[T] {
	_c.Call.Return(run)
	return _c
}

// NewBucketEntry creates a new instance of BucketEntry. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBucketEntry[T any](t interface {
	mock.TestingT
	Cleanup(func())
}) *BucketEntry[T] {
	mock := &BucketEntry[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
