// Code generated by mockery. DO NOT EDIT.

package mockapp

import (
	context "context"

	storage "github.com/adwski/shorty/internal/storage"
	mock "github.com/stretchr/testify/mock"
)

// Storage is an autogenerated mock type for the Storage type
type Storage struct {
	mock.Mock
}

type Storage_Expecter struct {
	mock *mock.Mock
}

func (_m *Storage) EXPECT() *Storage_Expecter {
	return &Storage_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *Storage) Close() {
	_m.Called()
}

// Storage_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type Storage_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *Storage_Expecter) Close() *Storage_Close_Call {
	return &Storage_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *Storage_Close_Call) Run(run func()) *Storage_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Storage_Close_Call) Return() *Storage_Close_Call {
	_c.Call.Return()
	return _c
}

func (_c *Storage_Close_Call) RunAndReturn(run func()) *Storage_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, key
func (_m *Storage) Get(ctx context.Context, key string) (string, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type Storage_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *Storage_Expecter) Get(ctx interface{}, key interface{}) *Storage_Get_Call {
	return &Storage_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *Storage_Get_Call) Run(run func(ctx context.Context, key string)) *Storage_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Storage_Get_Call) Return(url string, err error) *Storage_Get_Call {
	_c.Call.Return(url, err)
	return _c
}

func (_c *Storage_Get_Call) RunAndReturn(run func(context.Context, string) (string, error)) *Storage_Get_Call {
	_c.Call.Return(run)
	return _c
}

// ListUserURLs provides a mock function with given fields: ctx, uid
func (_m *Storage) ListUserURLs(ctx context.Context, uid string) ([]*storage.URL, error) {
	ret := _m.Called(ctx, uid)

	if len(ret) == 0 {
		panic("no return value specified for ListUserURLs")
	}

	var r0 []*storage.URL
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*storage.URL, error)); ok {
		return rf(ctx, uid)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*storage.URL); ok {
		r0 = rf(ctx, uid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*storage.URL)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, uid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_ListUserURLs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListUserURLs'
type Storage_ListUserURLs_Call struct {
	*mock.Call
}

// ListUserURLs is a helper method to define mock.On call
//   - ctx context.Context
//   - uid string
func (_e *Storage_Expecter) ListUserURLs(ctx interface{}, uid interface{}) *Storage_ListUserURLs_Call {
	return &Storage_ListUserURLs_Call{Call: _e.mock.On("ListUserURLs", ctx, uid)}
}

func (_c *Storage_ListUserURLs_Call) Run(run func(ctx context.Context, uid string)) *Storage_ListUserURLs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Storage_ListUserURLs_Call) Return(_a0 []*storage.URL, _a1 error) *Storage_ListUserURLs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_ListUserURLs_Call) RunAndReturn(run func(context.Context, string) ([]*storage.URL, error)) *Storage_ListUserURLs_Call {
	_c.Call.Return(run)
	return _c
}

// Ping provides a mock function with given fields: ctx
func (_m *Storage) Ping(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Ping")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_Ping_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Ping'
type Storage_Ping_Call struct {
	*mock.Call
}

// Ping is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Storage_Expecter) Ping(ctx interface{}) *Storage_Ping_Call {
	return &Storage_Ping_Call{Call: _e.mock.On("Ping", ctx)}
}

func (_c *Storage_Ping_Call) Run(run func(ctx context.Context)) *Storage_Ping_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Storage_Ping_Call) Return(_a0 error) *Storage_Ping_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_Ping_Call) RunAndReturn(run func(context.Context) error) *Storage_Ping_Call {
	_c.Call.Return(run)
	return _c
}

// Store provides a mock function with given fields: ctx, url, overwrite
func (_m *Storage) Store(ctx context.Context, url *storage.URL, overwrite bool) (string, error) {
	ret := _m.Called(ctx, url, overwrite)

	if len(ret) == 0 {
		panic("no return value specified for Store")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *storage.URL, bool) (string, error)); ok {
		return rf(ctx, url, overwrite)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *storage.URL, bool) string); ok {
		r0 = rf(ctx, url, overwrite)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *storage.URL, bool) error); ok {
		r1 = rf(ctx, url, overwrite)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_Store_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Store'
type Storage_Store_Call struct {
	*mock.Call
}

// Store is a helper method to define mock.On call
//   - ctx context.Context
//   - url *storage.URL
//   - overwrite bool
func (_e *Storage_Expecter) Store(ctx interface{}, url interface{}, overwrite interface{}) *Storage_Store_Call {
	return &Storage_Store_Call{Call: _e.mock.On("Store", ctx, url, overwrite)}
}

func (_c *Storage_Store_Call) Run(run func(ctx context.Context, url *storage.URL, overwrite bool)) *Storage_Store_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*storage.URL), args[2].(bool))
	})
	return _c
}

func (_c *Storage_Store_Call) Return(_a0 string, _a1 error) *Storage_Store_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_Store_Call) RunAndReturn(run func(context.Context, *storage.URL, bool) (string, error)) *Storage_Store_Call {
	_c.Call.Return(run)
	return _c
}

// StoreBatch provides a mock function with given fields: ctx, urls
func (_m *Storage) StoreBatch(ctx context.Context, urls []storage.URL) error {
	ret := _m.Called(ctx, urls)

	if len(ret) == 0 {
		panic("no return value specified for StoreBatch")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []storage.URL) error); ok {
		r0 = rf(ctx, urls)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_StoreBatch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StoreBatch'
type Storage_StoreBatch_Call struct {
	*mock.Call
}

// StoreBatch is a helper method to define mock.On call
//   - ctx context.Context
//   - urls []storage.URL
func (_e *Storage_Expecter) StoreBatch(ctx interface{}, urls interface{}) *Storage_StoreBatch_Call {
	return &Storage_StoreBatch_Call{Call: _e.mock.On("StoreBatch", ctx, urls)}
}

func (_c *Storage_StoreBatch_Call) Run(run func(ctx context.Context, urls []storage.URL)) *Storage_StoreBatch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]storage.URL))
	})
	return _c
}

func (_c *Storage_StoreBatch_Call) Return(_a0 error) *Storage_StoreBatch_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_StoreBatch_Call) RunAndReturn(run func(context.Context, []storage.URL) error) *Storage_StoreBatch_Call {
	_c.Call.Return(run)
	return _c
}

// NewStorage creates a new instance of Storage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *Storage {
	mock := &Storage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
