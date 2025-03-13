// Code generated by mockery. DO NOT EDIT.

package peanats_test

import (
	context "context"

	jetstream "github.com/nats-io/nats.go/jetstream"
	mock "github.com/stretchr/testify/mock"
)

// keyValueMock is an autogenerated mock type for the KeyValue type
type keyValueMock struct {
	mock.Mock
}

type keyValueMock_Expecter struct {
	mock *mock.Mock
}

func (_m *keyValueMock) EXPECT() *keyValueMock_Expecter {
	return &keyValueMock_Expecter{mock: &_m.Mock}
}

// Bucket provides a mock function with no fields
func (_m *keyValueMock) Bucket() string {
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

// keyValueMock_Bucket_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Bucket'
type keyValueMock_Bucket_Call struct {
	*mock.Call
}

// Bucket is a helper method to define mock.On call
func (_e *keyValueMock_Expecter) Bucket() *keyValueMock_Bucket_Call {
	return &keyValueMock_Bucket_Call{Call: _e.mock.On("Bucket")}
}

func (_c *keyValueMock_Bucket_Call) Run(run func()) *keyValueMock_Bucket_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *keyValueMock_Bucket_Call) Return(_a0 string) *keyValueMock_Bucket_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *keyValueMock_Bucket_Call) RunAndReturn(run func() string) *keyValueMock_Bucket_Call {
	_c.Call.Return(run)
	return _c
}

// Create provides a mock function with given fields: ctx, key, value
func (_m *keyValueMock) Create(ctx context.Context, key string, value []byte) (uint64, error) {
	ret := _m.Called(ctx, key, value)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []byte) (uint64, error)); ok {
		return rf(ctx, key, value)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []byte) uint64); ok {
		r0 = rf(ctx, key, value)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []byte) error); ok {
		r1 = rf(ctx, key, value)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type keyValueMock_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - value []byte
func (_e *keyValueMock_Expecter) Create(ctx interface{}, key interface{}, value interface{}) *keyValueMock_Create_Call {
	return &keyValueMock_Create_Call{Call: _e.mock.On("Create", ctx, key, value)}
}

func (_c *keyValueMock_Create_Call) Run(run func(ctx context.Context, key string, value []byte)) *keyValueMock_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].([]byte))
	})
	return _c
}

func (_c *keyValueMock_Create_Call) Return(_a0 uint64, _a1 error) *keyValueMock_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_Create_Call) RunAndReturn(run func(context.Context, string, []byte) (uint64, error)) *keyValueMock_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: ctx, key, opts
func (_m *keyValueMock) Delete(ctx context.Context, key string, opts ...jetstream.KVDeleteOpt) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, key)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...jetstream.KVDeleteOpt) error); ok {
		r0 = rf(ctx, key, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// keyValueMock_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type keyValueMock_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - opts ...jetstream.KVDeleteOpt
func (_e *keyValueMock_Expecter) Delete(ctx interface{}, key interface{}, opts ...interface{}) *keyValueMock_Delete_Call {
	return &keyValueMock_Delete_Call{Call: _e.mock.On("Delete",
		append([]interface{}{ctx, key}, opts...)...)}
}

func (_c *keyValueMock_Delete_Call) Run(run func(ctx context.Context, key string, opts ...jetstream.KVDeleteOpt)) *keyValueMock_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]jetstream.KVDeleteOpt, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(jetstream.KVDeleteOpt)
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *keyValueMock_Delete_Call) Return(_a0 error) *keyValueMock_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *keyValueMock_Delete_Call) RunAndReturn(run func(context.Context, string, ...jetstream.KVDeleteOpt) error) *keyValueMock_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, key
func (_m *keyValueMock) Get(ctx context.Context, key string) (jetstream.KeyValueEntry, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 jetstream.KeyValueEntry
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (jetstream.KeyValueEntry, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) jetstream.KeyValueEntry); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.KeyValueEntry)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type keyValueMock_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *keyValueMock_Expecter) Get(ctx interface{}, key interface{}) *keyValueMock_Get_Call {
	return &keyValueMock_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *keyValueMock_Get_Call) Run(run func(ctx context.Context, key string)) *keyValueMock_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *keyValueMock_Get_Call) Return(_a0 jetstream.KeyValueEntry, _a1 error) *keyValueMock_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_Get_Call) RunAndReturn(run func(context.Context, string) (jetstream.KeyValueEntry, error)) *keyValueMock_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetRevision provides a mock function with given fields: ctx, key, revision
func (_m *keyValueMock) GetRevision(ctx context.Context, key string, revision uint64) (jetstream.KeyValueEntry, error) {
	ret := _m.Called(ctx, key, revision)

	if len(ret) == 0 {
		panic("no return value specified for GetRevision")
	}

	var r0 jetstream.KeyValueEntry
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uint64) (jetstream.KeyValueEntry, error)); ok {
		return rf(ctx, key, revision)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, uint64) jetstream.KeyValueEntry); ok {
		r0 = rf(ctx, key, revision)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.KeyValueEntry)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, uint64) error); ok {
		r1 = rf(ctx, key, revision)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_GetRevision_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRevision'
type keyValueMock_GetRevision_Call struct {
	*mock.Call
}

// GetRevision is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - revision uint64
func (_e *keyValueMock_Expecter) GetRevision(ctx interface{}, key interface{}, revision interface{}) *keyValueMock_GetRevision_Call {
	return &keyValueMock_GetRevision_Call{Call: _e.mock.On("GetRevision", ctx, key, revision)}
}

func (_c *keyValueMock_GetRevision_Call) Run(run func(ctx context.Context, key string, revision uint64)) *keyValueMock_GetRevision_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(uint64))
	})
	return _c
}

func (_c *keyValueMock_GetRevision_Call) Return(_a0 jetstream.KeyValueEntry, _a1 error) *keyValueMock_GetRevision_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_GetRevision_Call) RunAndReturn(run func(context.Context, string, uint64) (jetstream.KeyValueEntry, error)) *keyValueMock_GetRevision_Call {
	_c.Call.Return(run)
	return _c
}

// History provides a mock function with given fields: ctx, key, opts
func (_m *keyValueMock) History(ctx context.Context, key string, opts ...jetstream.WatchOpt) ([]jetstream.KeyValueEntry, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, key)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for History")
	}

	var r0 []jetstream.KeyValueEntry
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...jetstream.WatchOpt) ([]jetstream.KeyValueEntry, error)); ok {
		return rf(ctx, key, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...jetstream.WatchOpt) []jetstream.KeyValueEntry); ok {
		r0 = rf(ctx, key, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]jetstream.KeyValueEntry)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...jetstream.WatchOpt) error); ok {
		r1 = rf(ctx, key, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_History_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'History'
type keyValueMock_History_Call struct {
	*mock.Call
}

// History is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - opts ...jetstream.WatchOpt
func (_e *keyValueMock_Expecter) History(ctx interface{}, key interface{}, opts ...interface{}) *keyValueMock_History_Call {
	return &keyValueMock_History_Call{Call: _e.mock.On("History",
		append([]interface{}{ctx, key}, opts...)...)}
}

func (_c *keyValueMock_History_Call) Run(run func(ctx context.Context, key string, opts ...jetstream.WatchOpt)) *keyValueMock_History_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]jetstream.WatchOpt, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(jetstream.WatchOpt)
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *keyValueMock_History_Call) Return(_a0 []jetstream.KeyValueEntry, _a1 error) *keyValueMock_History_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_History_Call) RunAndReturn(run func(context.Context, string, ...jetstream.WatchOpt) ([]jetstream.KeyValueEntry, error)) *keyValueMock_History_Call {
	_c.Call.Return(run)
	return _c
}

// Keys provides a mock function with given fields: ctx, opts
func (_m *keyValueMock) Keys(ctx context.Context, opts ...jetstream.WatchOpt) ([]string, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Keys")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...jetstream.WatchOpt) ([]string, error)); ok {
		return rf(ctx, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...jetstream.WatchOpt) []string); ok {
		r0 = rf(ctx, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...jetstream.WatchOpt) error); ok {
		r1 = rf(ctx, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_Keys_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Keys'
type keyValueMock_Keys_Call struct {
	*mock.Call
}

// Keys is a helper method to define mock.On call
//   - ctx context.Context
//   - opts ...jetstream.WatchOpt
func (_e *keyValueMock_Expecter) Keys(ctx interface{}, opts ...interface{}) *keyValueMock_Keys_Call {
	return &keyValueMock_Keys_Call{Call: _e.mock.On("Keys",
		append([]interface{}{ctx}, opts...)...)}
}

func (_c *keyValueMock_Keys_Call) Run(run func(ctx context.Context, opts ...jetstream.WatchOpt)) *keyValueMock_Keys_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]jetstream.WatchOpt, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(jetstream.WatchOpt)
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *keyValueMock_Keys_Call) Return(_a0 []string, _a1 error) *keyValueMock_Keys_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_Keys_Call) RunAndReturn(run func(context.Context, ...jetstream.WatchOpt) ([]string, error)) *keyValueMock_Keys_Call {
	_c.Call.Return(run)
	return _c
}

// ListKeys provides a mock function with given fields: ctx, opts
func (_m *keyValueMock) ListKeys(ctx context.Context, opts ...jetstream.WatchOpt) (jetstream.KeyLister, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListKeys")
	}

	var r0 jetstream.KeyLister
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...jetstream.WatchOpt) (jetstream.KeyLister, error)); ok {
		return rf(ctx, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...jetstream.WatchOpt) jetstream.KeyLister); ok {
		r0 = rf(ctx, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.KeyLister)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...jetstream.WatchOpt) error); ok {
		r1 = rf(ctx, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_ListKeys_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListKeys'
type keyValueMock_ListKeys_Call struct {
	*mock.Call
}

// ListKeys is a helper method to define mock.On call
//   - ctx context.Context
//   - opts ...jetstream.WatchOpt
func (_e *keyValueMock_Expecter) ListKeys(ctx interface{}, opts ...interface{}) *keyValueMock_ListKeys_Call {
	return &keyValueMock_ListKeys_Call{Call: _e.mock.On("ListKeys",
		append([]interface{}{ctx}, opts...)...)}
}

func (_c *keyValueMock_ListKeys_Call) Run(run func(ctx context.Context, opts ...jetstream.WatchOpt)) *keyValueMock_ListKeys_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]jetstream.WatchOpt, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(jetstream.WatchOpt)
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *keyValueMock_ListKeys_Call) Return(_a0 jetstream.KeyLister, _a1 error) *keyValueMock_ListKeys_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_ListKeys_Call) RunAndReturn(run func(context.Context, ...jetstream.WatchOpt) (jetstream.KeyLister, error)) *keyValueMock_ListKeys_Call {
	_c.Call.Return(run)
	return _c
}

// ListKeysFiltered provides a mock function with given fields: ctx, filters
func (_m *keyValueMock) ListKeysFiltered(ctx context.Context, filters ...string) (jetstream.KeyLister, error) {
	_va := make([]interface{}, len(filters))
	for _i := range filters {
		_va[_i] = filters[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListKeysFiltered")
	}

	var r0 jetstream.KeyLister
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...string) (jetstream.KeyLister, error)); ok {
		return rf(ctx, filters...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...string) jetstream.KeyLister); ok {
		r0 = rf(ctx, filters...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.KeyLister)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...string) error); ok {
		r1 = rf(ctx, filters...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_ListKeysFiltered_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListKeysFiltered'
type keyValueMock_ListKeysFiltered_Call struct {
	*mock.Call
}

// ListKeysFiltered is a helper method to define mock.On call
//   - ctx context.Context
//   - filters ...string
func (_e *keyValueMock_Expecter) ListKeysFiltered(ctx interface{}, filters ...interface{}) *keyValueMock_ListKeysFiltered_Call {
	return &keyValueMock_ListKeysFiltered_Call{Call: _e.mock.On("ListKeysFiltered",
		append([]interface{}{ctx}, filters...)...)}
}

func (_c *keyValueMock_ListKeysFiltered_Call) Run(run func(ctx context.Context, filters ...string)) *keyValueMock_ListKeysFiltered_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *keyValueMock_ListKeysFiltered_Call) Return(_a0 jetstream.KeyLister, _a1 error) *keyValueMock_ListKeysFiltered_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_ListKeysFiltered_Call) RunAndReturn(run func(context.Context, ...string) (jetstream.KeyLister, error)) *keyValueMock_ListKeysFiltered_Call {
	_c.Call.Return(run)
	return _c
}

// Purge provides a mock function with given fields: ctx, key, opts
func (_m *keyValueMock) Purge(ctx context.Context, key string, opts ...jetstream.KVDeleteOpt) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, key)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Purge")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...jetstream.KVDeleteOpt) error); ok {
		r0 = rf(ctx, key, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// keyValueMock_Purge_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Purge'
type keyValueMock_Purge_Call struct {
	*mock.Call
}

// Purge is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - opts ...jetstream.KVDeleteOpt
func (_e *keyValueMock_Expecter) Purge(ctx interface{}, key interface{}, opts ...interface{}) *keyValueMock_Purge_Call {
	return &keyValueMock_Purge_Call{Call: _e.mock.On("Purge",
		append([]interface{}{ctx, key}, opts...)...)}
}

func (_c *keyValueMock_Purge_Call) Run(run func(ctx context.Context, key string, opts ...jetstream.KVDeleteOpt)) *keyValueMock_Purge_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]jetstream.KVDeleteOpt, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(jetstream.KVDeleteOpt)
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *keyValueMock_Purge_Call) Return(_a0 error) *keyValueMock_Purge_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *keyValueMock_Purge_Call) RunAndReturn(run func(context.Context, string, ...jetstream.KVDeleteOpt) error) *keyValueMock_Purge_Call {
	_c.Call.Return(run)
	return _c
}

// PurgeDeletes provides a mock function with given fields: ctx, opts
func (_m *keyValueMock) PurgeDeletes(ctx context.Context, opts ...jetstream.KVPurgeOpt) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for PurgeDeletes")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...jetstream.KVPurgeOpt) error); ok {
		r0 = rf(ctx, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// keyValueMock_PurgeDeletes_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PurgeDeletes'
type keyValueMock_PurgeDeletes_Call struct {
	*mock.Call
}

// PurgeDeletes is a helper method to define mock.On call
//   - ctx context.Context
//   - opts ...jetstream.KVPurgeOpt
func (_e *keyValueMock_Expecter) PurgeDeletes(ctx interface{}, opts ...interface{}) *keyValueMock_PurgeDeletes_Call {
	return &keyValueMock_PurgeDeletes_Call{Call: _e.mock.On("PurgeDeletes",
		append([]interface{}{ctx}, opts...)...)}
}

func (_c *keyValueMock_PurgeDeletes_Call) Run(run func(ctx context.Context, opts ...jetstream.KVPurgeOpt)) *keyValueMock_PurgeDeletes_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]jetstream.KVPurgeOpt, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(jetstream.KVPurgeOpt)
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *keyValueMock_PurgeDeletes_Call) Return(_a0 error) *keyValueMock_PurgeDeletes_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *keyValueMock_PurgeDeletes_Call) RunAndReturn(run func(context.Context, ...jetstream.KVPurgeOpt) error) *keyValueMock_PurgeDeletes_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: ctx, key, value
func (_m *keyValueMock) Put(ctx context.Context, key string, value []byte) (uint64, error) {
	ret := _m.Called(ctx, key, value)

	if len(ret) == 0 {
		panic("no return value specified for Put")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []byte) (uint64, error)); ok {
		return rf(ctx, key, value)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []byte) uint64); ok {
		r0 = rf(ctx, key, value)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []byte) error); ok {
		r1 = rf(ctx, key, value)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type keyValueMock_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - value []byte
func (_e *keyValueMock_Expecter) Put(ctx interface{}, key interface{}, value interface{}) *keyValueMock_Put_Call {
	return &keyValueMock_Put_Call{Call: _e.mock.On("Put", ctx, key, value)}
}

func (_c *keyValueMock_Put_Call) Run(run func(ctx context.Context, key string, value []byte)) *keyValueMock_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].([]byte))
	})
	return _c
}

func (_c *keyValueMock_Put_Call) Return(_a0 uint64, _a1 error) *keyValueMock_Put_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_Put_Call) RunAndReturn(run func(context.Context, string, []byte) (uint64, error)) *keyValueMock_Put_Call {
	_c.Call.Return(run)
	return _c
}

// PutString provides a mock function with given fields: ctx, key, value
func (_m *keyValueMock) PutString(ctx context.Context, key string, value string) (uint64, error) {
	ret := _m.Called(ctx, key, value)

	if len(ret) == 0 {
		panic("no return value specified for PutString")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (uint64, error)); ok {
		return rf(ctx, key, value)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) uint64); ok {
		r0 = rf(ctx, key, value)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, key, value)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_PutString_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PutString'
type keyValueMock_PutString_Call struct {
	*mock.Call
}

// PutString is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - value string
func (_e *keyValueMock_Expecter) PutString(ctx interface{}, key interface{}, value interface{}) *keyValueMock_PutString_Call {
	return &keyValueMock_PutString_Call{Call: _e.mock.On("PutString", ctx, key, value)}
}

func (_c *keyValueMock_PutString_Call) Run(run func(ctx context.Context, key string, value string)) *keyValueMock_PutString_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *keyValueMock_PutString_Call) Return(_a0 uint64, _a1 error) *keyValueMock_PutString_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_PutString_Call) RunAndReturn(run func(context.Context, string, string) (uint64, error)) *keyValueMock_PutString_Call {
	_c.Call.Return(run)
	return _c
}

// Status provides a mock function with given fields: ctx
func (_m *keyValueMock) Status(ctx context.Context) (jetstream.KeyValueStatus, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Status")
	}

	var r0 jetstream.KeyValueStatus
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (jetstream.KeyValueStatus, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) jetstream.KeyValueStatus); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.KeyValueStatus)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_Status_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Status'
type keyValueMock_Status_Call struct {
	*mock.Call
}

// Status is a helper method to define mock.On call
//   - ctx context.Context
func (_e *keyValueMock_Expecter) Status(ctx interface{}) *keyValueMock_Status_Call {
	return &keyValueMock_Status_Call{Call: _e.mock.On("Status", ctx)}
}

func (_c *keyValueMock_Status_Call) Run(run func(ctx context.Context)) *keyValueMock_Status_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *keyValueMock_Status_Call) Return(_a0 jetstream.KeyValueStatus, _a1 error) *keyValueMock_Status_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_Status_Call) RunAndReturn(run func(context.Context) (jetstream.KeyValueStatus, error)) *keyValueMock_Status_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, key, value, revision
func (_m *keyValueMock) Update(ctx context.Context, key string, value []byte, revision uint64) (uint64, error) {
	ret := _m.Called(ctx, key, value, revision)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []byte, uint64) (uint64, error)); ok {
		return rf(ctx, key, value, revision)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []byte, uint64) uint64); ok {
		r0 = rf(ctx, key, value, revision)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []byte, uint64) error); ok {
		r1 = rf(ctx, key, value, revision)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type keyValueMock_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - value []byte
//   - revision uint64
func (_e *keyValueMock_Expecter) Update(ctx interface{}, key interface{}, value interface{}, revision interface{}) *keyValueMock_Update_Call {
	return &keyValueMock_Update_Call{Call: _e.mock.On("Update", ctx, key, value, revision)}
}

func (_c *keyValueMock_Update_Call) Run(run func(ctx context.Context, key string, value []byte, revision uint64)) *keyValueMock_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].([]byte), args[3].(uint64))
	})
	return _c
}

func (_c *keyValueMock_Update_Call) Return(_a0 uint64, _a1 error) *keyValueMock_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_Update_Call) RunAndReturn(run func(context.Context, string, []byte, uint64) (uint64, error)) *keyValueMock_Update_Call {
	_c.Call.Return(run)
	return _c
}

// Watch provides a mock function with given fields: ctx, keys, opts
func (_m *keyValueMock) Watch(ctx context.Context, keys string, opts ...jetstream.WatchOpt) (jetstream.KeyWatcher, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, keys)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Watch")
	}

	var r0 jetstream.KeyWatcher
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...jetstream.WatchOpt) (jetstream.KeyWatcher, error)); ok {
		return rf(ctx, keys, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...jetstream.WatchOpt) jetstream.KeyWatcher); ok {
		r0 = rf(ctx, keys, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.KeyWatcher)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...jetstream.WatchOpt) error); ok {
		r1 = rf(ctx, keys, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_Watch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Watch'
type keyValueMock_Watch_Call struct {
	*mock.Call
}

// Watch is a helper method to define mock.On call
//   - ctx context.Context
//   - keys string
//   - opts ...jetstream.WatchOpt
func (_e *keyValueMock_Expecter) Watch(ctx interface{}, keys interface{}, opts ...interface{}) *keyValueMock_Watch_Call {
	return &keyValueMock_Watch_Call{Call: _e.mock.On("Watch",
		append([]interface{}{ctx, keys}, opts...)...)}
}

func (_c *keyValueMock_Watch_Call) Run(run func(ctx context.Context, keys string, opts ...jetstream.WatchOpt)) *keyValueMock_Watch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]jetstream.WatchOpt, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(jetstream.WatchOpt)
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *keyValueMock_Watch_Call) Return(_a0 jetstream.KeyWatcher, _a1 error) *keyValueMock_Watch_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_Watch_Call) RunAndReturn(run func(context.Context, string, ...jetstream.WatchOpt) (jetstream.KeyWatcher, error)) *keyValueMock_Watch_Call {
	_c.Call.Return(run)
	return _c
}

// WatchAll provides a mock function with given fields: ctx, opts
func (_m *keyValueMock) WatchAll(ctx context.Context, opts ...jetstream.WatchOpt) (jetstream.KeyWatcher, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for WatchAll")
	}

	var r0 jetstream.KeyWatcher
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...jetstream.WatchOpt) (jetstream.KeyWatcher, error)); ok {
		return rf(ctx, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...jetstream.WatchOpt) jetstream.KeyWatcher); ok {
		r0 = rf(ctx, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.KeyWatcher)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...jetstream.WatchOpt) error); ok {
		r1 = rf(ctx, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_WatchAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WatchAll'
type keyValueMock_WatchAll_Call struct {
	*mock.Call
}

// WatchAll is a helper method to define mock.On call
//   - ctx context.Context
//   - opts ...jetstream.WatchOpt
func (_e *keyValueMock_Expecter) WatchAll(ctx interface{}, opts ...interface{}) *keyValueMock_WatchAll_Call {
	return &keyValueMock_WatchAll_Call{Call: _e.mock.On("WatchAll",
		append([]interface{}{ctx}, opts...)...)}
}

func (_c *keyValueMock_WatchAll_Call) Run(run func(ctx context.Context, opts ...jetstream.WatchOpt)) *keyValueMock_WatchAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]jetstream.WatchOpt, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(jetstream.WatchOpt)
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *keyValueMock_WatchAll_Call) Return(_a0 jetstream.KeyWatcher, _a1 error) *keyValueMock_WatchAll_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_WatchAll_Call) RunAndReturn(run func(context.Context, ...jetstream.WatchOpt) (jetstream.KeyWatcher, error)) *keyValueMock_WatchAll_Call {
	_c.Call.Return(run)
	return _c
}

// WatchFiltered provides a mock function with given fields: ctx, keys, opts
func (_m *keyValueMock) WatchFiltered(ctx context.Context, keys []string, opts ...jetstream.WatchOpt) (jetstream.KeyWatcher, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, keys)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for WatchFiltered")
	}

	var r0 jetstream.KeyWatcher
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string, ...jetstream.WatchOpt) (jetstream.KeyWatcher, error)); ok {
		return rf(ctx, keys, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string, ...jetstream.WatchOpt) jetstream.KeyWatcher); ok {
		r0 = rf(ctx, keys, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetstream.KeyWatcher)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string, ...jetstream.WatchOpt) error); ok {
		r1 = rf(ctx, keys, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// keyValueMock_WatchFiltered_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WatchFiltered'
type keyValueMock_WatchFiltered_Call struct {
	*mock.Call
}

// WatchFiltered is a helper method to define mock.On call
//   - ctx context.Context
//   - keys []string
//   - opts ...jetstream.WatchOpt
func (_e *keyValueMock_Expecter) WatchFiltered(ctx interface{}, keys interface{}, opts ...interface{}) *keyValueMock_WatchFiltered_Call {
	return &keyValueMock_WatchFiltered_Call{Call: _e.mock.On("WatchFiltered",
		append([]interface{}{ctx, keys}, opts...)...)}
}

func (_c *keyValueMock_WatchFiltered_Call) Run(run func(ctx context.Context, keys []string, opts ...jetstream.WatchOpt)) *keyValueMock_WatchFiltered_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]jetstream.WatchOpt, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(jetstream.WatchOpt)
			}
		}
		run(args[0].(context.Context), args[1].([]string), variadicArgs...)
	})
	return _c
}

func (_c *keyValueMock_WatchFiltered_Call) Return(_a0 jetstream.KeyWatcher, _a1 error) *keyValueMock_WatchFiltered_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *keyValueMock_WatchFiltered_Call) RunAndReturn(run func(context.Context, []string, ...jetstream.WatchOpt) (jetstream.KeyWatcher, error)) *keyValueMock_WatchFiltered_Call {
	_c.Call.Return(run)
	return _c
}

// newKeyValueMock creates a new instance of keyValueMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newKeyValueMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *keyValueMock {
	mock := &keyValueMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
