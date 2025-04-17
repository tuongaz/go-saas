// Code generated by mockery v2.53.3. DO NOT EDIT.

package store

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	store "github.com/tuongaz/go-saas/store"

	types "github.com/tuongaz/go-saas/store/types"
)

// MockCollectionInterface is an autogenerated mock type for the CollectionInterface type
type MockCollectionInterface struct {
	mock.Mock
}

type MockCollectionInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCollectionInterface) EXPECT() *MockCollectionInterface_Expecter {
	return &MockCollectionInterface_Expecter{mock: &_m.Mock}
}

// Count provides a mock function with given fields: ctx, filter
func (_m *MockCollectionInterface) Count(ctx context.Context, filter store.Filter) (int, error) {
	ret := _m.Called(ctx, filter)

	if len(ret) == 0 {
		panic("no return value specified for Count")
	}

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, store.Filter) (int, error)); ok {
		return rf(ctx, filter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, store.Filter) int); ok {
		r0 = rf(ctx, filter)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(context.Context, store.Filter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCollectionInterface_Count_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Count'
type MockCollectionInterface_Count_Call struct {
	*mock.Call
}

// Count is a helper method to define mock.On call
//   - ctx context.Context
//   - filter store.Filter
func (_e *MockCollectionInterface_Expecter) Count(ctx interface{}, filter interface{}) *MockCollectionInterface_Count_Call {
	return &MockCollectionInterface_Count_Call{Call: _e.mock.On("Count", ctx, filter)}
}

func (_c *MockCollectionInterface_Count_Call) Run(run func(ctx context.Context, filter store.Filter)) *MockCollectionInterface_Count_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(store.Filter))
	})
	return _c
}

func (_c *MockCollectionInterface_Count_Call) Return(_a0 int, _a1 error) *MockCollectionInterface_Count_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCollectionInterface_Count_Call) RunAndReturn(run func(context.Context, store.Filter) (int, error)) *MockCollectionInterface_Count_Call {
	_c.Call.Return(run)
	return _c
}

// CreateRecord provides a mock function with given fields: ctx, record
func (_m *MockCollectionInterface) CreateRecord(ctx context.Context, record types.Record) (*types.Record, error) {
	ret := _m.Called(ctx, record)

	if len(ret) == 0 {
		panic("no return value specified for CreateRecord")
	}

	var r0 *types.Record
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, types.Record) (*types.Record, error)); ok {
		return rf(ctx, record)
	}
	if rf, ok := ret.Get(0).(func(context.Context, types.Record) *types.Record); ok {
		r0 = rf(ctx, record)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Record)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, types.Record) error); ok {
		r1 = rf(ctx, record)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCollectionInterface_CreateRecord_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateRecord'
type MockCollectionInterface_CreateRecord_Call struct {
	*mock.Call
}

// CreateRecord is a helper method to define mock.On call
//   - ctx context.Context
//   - record types.Record
func (_e *MockCollectionInterface_Expecter) CreateRecord(ctx interface{}, record interface{}) *MockCollectionInterface_CreateRecord_Call {
	return &MockCollectionInterface_CreateRecord_Call{Call: _e.mock.On("CreateRecord", ctx, record)}
}

func (_c *MockCollectionInterface_CreateRecord_Call) Run(run func(ctx context.Context, record types.Record)) *MockCollectionInterface_CreateRecord_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(types.Record))
	})
	return _c
}

func (_c *MockCollectionInterface_CreateRecord_Call) Return(_a0 *types.Record, _a1 error) *MockCollectionInterface_CreateRecord_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCollectionInterface_CreateRecord_Call) RunAndReturn(run func(context.Context, types.Record) (*types.Record, error)) *MockCollectionInterface_CreateRecord_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteRecord provides a mock function with given fields: ctx, id
func (_m *MockCollectionInterface) DeleteRecord(ctx context.Context, id interface{}) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteRecord")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCollectionInterface_DeleteRecord_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteRecord'
type MockCollectionInterface_DeleteRecord_Call struct {
	*mock.Call
}

// DeleteRecord is a helper method to define mock.On call
//   - ctx context.Context
//   - id interface{}
func (_e *MockCollectionInterface_Expecter) DeleteRecord(ctx interface{}, id interface{}) *MockCollectionInterface_DeleteRecord_Call {
	return &MockCollectionInterface_DeleteRecord_Call{Call: _e.mock.On("DeleteRecord", ctx, id)}
}

func (_c *MockCollectionInterface_DeleteRecord_Call) Run(run func(ctx context.Context, id interface{})) *MockCollectionInterface_DeleteRecord_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interface{}))
	})
	return _c
}

func (_c *MockCollectionInterface_DeleteRecord_Call) Return(_a0 error) *MockCollectionInterface_DeleteRecord_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCollectionInterface_DeleteRecord_Call) RunAndReturn(run func(context.Context, interface{}) error) *MockCollectionInterface_DeleteRecord_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteRecords provides a mock function with given fields: ctx, filter
func (_m *MockCollectionInterface) DeleteRecords(ctx context.Context, filter store.Filter) error {
	ret := _m.Called(ctx, filter)

	if len(ret) == 0 {
		panic("no return value specified for DeleteRecords")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, store.Filter) error); ok {
		r0 = rf(ctx, filter)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCollectionInterface_DeleteRecords_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteRecords'
type MockCollectionInterface_DeleteRecords_Call struct {
	*mock.Call
}

// DeleteRecords is a helper method to define mock.On call
//   - ctx context.Context
//   - filter store.Filter
func (_e *MockCollectionInterface_Expecter) DeleteRecords(ctx interface{}, filter interface{}) *MockCollectionInterface_DeleteRecords_Call {
	return &MockCollectionInterface_DeleteRecords_Call{Call: _e.mock.On("DeleteRecords", ctx, filter)}
}

func (_c *MockCollectionInterface_DeleteRecords_Call) Run(run func(ctx context.Context, filter store.Filter)) *MockCollectionInterface_DeleteRecords_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(store.Filter))
	})
	return _c
}

func (_c *MockCollectionInterface_DeleteRecords_Call) Return(_a0 error) *MockCollectionInterface_DeleteRecords_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCollectionInterface_DeleteRecords_Call) RunAndReturn(run func(context.Context, store.Filter) error) *MockCollectionInterface_DeleteRecords_Call {
	_c.Call.Return(run)
	return _c
}

// Exists provides a mock function with given fields: ctx, filter
func (_m *MockCollectionInterface) Exists(ctx context.Context, filter store.Filter) (bool, error) {
	ret := _m.Called(ctx, filter)

	if len(ret) == 0 {
		panic("no return value specified for Exists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, store.Filter) (bool, error)); ok {
		return rf(ctx, filter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, store.Filter) bool); ok {
		r0 = rf(ctx, filter)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, store.Filter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCollectionInterface_Exists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Exists'
type MockCollectionInterface_Exists_Call struct {
	*mock.Call
}

// Exists is a helper method to define mock.On call
//   - ctx context.Context
//   - filter store.Filter
func (_e *MockCollectionInterface_Expecter) Exists(ctx interface{}, filter interface{}) *MockCollectionInterface_Exists_Call {
	return &MockCollectionInterface_Exists_Call{Call: _e.mock.On("Exists", ctx, filter)}
}

func (_c *MockCollectionInterface_Exists_Call) Run(run func(ctx context.Context, filter store.Filter)) *MockCollectionInterface_Exists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(store.Filter))
	})
	return _c
}

func (_c *MockCollectionInterface_Exists_Call) Return(_a0 bool, _a1 error) *MockCollectionInterface_Exists_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCollectionInterface_Exists_Call) RunAndReturn(run func(context.Context, store.Filter) (bool, error)) *MockCollectionInterface_Exists_Call {
	_c.Call.Return(run)
	return _c
}

// Find provides a mock function with given fields: ctx, opts
func (_m *MockCollectionInterface) Find(ctx context.Context, opts ...store.FindOption) (*store.List, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Find")
	}

	var r0 *store.List
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...store.FindOption) (*store.List, error)); ok {
		return rf(ctx, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...store.FindOption) *store.List); ok {
		r0 = rf(ctx, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*store.List)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...store.FindOption) error); ok {
		r1 = rf(ctx, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCollectionInterface_Find_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Find'
type MockCollectionInterface_Find_Call struct {
	*mock.Call
}

// Find is a helper method to define mock.On call
//   - ctx context.Context
//   - opts ...store.FindOption
func (_e *MockCollectionInterface_Expecter) Find(ctx interface{}, opts ...interface{}) *MockCollectionInterface_Find_Call {
	return &MockCollectionInterface_Find_Call{Call: _e.mock.On("Find",
		append([]interface{}{ctx}, opts...)...)}
}

func (_c *MockCollectionInterface_Find_Call) Run(run func(ctx context.Context, opts ...store.FindOption)) *MockCollectionInterface_Find_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]store.FindOption, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(store.FindOption)
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *MockCollectionInterface_Find_Call) Return(_a0 *store.List, _a1 error) *MockCollectionInterface_Find_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCollectionInterface_Find_Call) RunAndReturn(run func(context.Context, ...store.FindOption) (*store.List, error)) *MockCollectionInterface_Find_Call {
	_c.Call.Return(run)
	return _c
}

// FindOne provides a mock function with given fields: ctx, filter
func (_m *MockCollectionInterface) FindOne(ctx context.Context, filter store.Filter) (*types.Record, error) {
	ret := _m.Called(ctx, filter)

	if len(ret) == 0 {
		panic("no return value specified for FindOne")
	}

	var r0 *types.Record
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, store.Filter) (*types.Record, error)); ok {
		return rf(ctx, filter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, store.Filter) *types.Record); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Record)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, store.Filter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCollectionInterface_FindOne_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FindOne'
type MockCollectionInterface_FindOne_Call struct {
	*mock.Call
}

// FindOne is a helper method to define mock.On call
//   - ctx context.Context
//   - filter store.Filter
func (_e *MockCollectionInterface_Expecter) FindOne(ctx interface{}, filter interface{}) *MockCollectionInterface_FindOne_Call {
	return &MockCollectionInterface_FindOne_Call{Call: _e.mock.On("FindOne", ctx, filter)}
}

func (_c *MockCollectionInterface_FindOne_Call) Run(run func(ctx context.Context, filter store.Filter)) *MockCollectionInterface_FindOne_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(store.Filter))
	})
	return _c
}

func (_c *MockCollectionInterface_FindOne_Call) Return(_a0 *types.Record, _a1 error) *MockCollectionInterface_FindOne_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCollectionInterface_FindOne_Call) RunAndReturn(run func(context.Context, store.Filter) (*types.Record, error)) *MockCollectionInterface_FindOne_Call {
	_c.Call.Return(run)
	return _c
}

// GetRecord provides a mock function with given fields: ctx, id
func (_m *MockCollectionInterface) GetRecord(ctx context.Context, id interface{}) (*types.Record, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetRecord")
	}

	var r0 *types.Record
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}) (*types.Record, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, interface{}) *types.Record); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Record)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, interface{}) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCollectionInterface_GetRecord_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRecord'
type MockCollectionInterface_GetRecord_Call struct {
	*mock.Call
}

// GetRecord is a helper method to define mock.On call
//   - ctx context.Context
//   - id interface{}
func (_e *MockCollectionInterface_Expecter) GetRecord(ctx interface{}, id interface{}) *MockCollectionInterface_GetRecord_Call {
	return &MockCollectionInterface_GetRecord_Call{Call: _e.mock.On("GetRecord", ctx, id)}
}

func (_c *MockCollectionInterface_GetRecord_Call) Run(run func(ctx context.Context, id interface{})) *MockCollectionInterface_GetRecord_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interface{}))
	})
	return _c
}

func (_c *MockCollectionInterface_GetRecord_Call) Return(_a0 *types.Record, _a1 error) *MockCollectionInterface_GetRecord_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCollectionInterface_GetRecord_Call) RunAndReturn(run func(context.Context, interface{}) (*types.Record, error)) *MockCollectionInterface_GetRecord_Call {
	_c.Call.Return(run)
	return _c
}

// Table provides a mock function with no fields
func (_m *MockCollectionInterface) Table() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Table")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockCollectionInterface_Table_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Table'
type MockCollectionInterface_Table_Call struct {
	*mock.Call
}

// Table is a helper method to define mock.On call
func (_e *MockCollectionInterface_Expecter) Table() *MockCollectionInterface_Table_Call {
	return &MockCollectionInterface_Table_Call{Call: _e.mock.On("Table")}
}

func (_c *MockCollectionInterface_Table_Call) Run(run func()) *MockCollectionInterface_Table_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockCollectionInterface_Table_Call) Return(_a0 string) *MockCollectionInterface_Table_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCollectionInterface_Table_Call) RunAndReturn(run func() string) *MockCollectionInterface_Table_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, record, args
func (_m *MockCollectionInterface) Update(ctx context.Context, record types.Record, args ...interface{}) (int64, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, record)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, types.Record, ...interface{}) (int64, error)); ok {
		return rf(ctx, record, args...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, types.Record, ...interface{}) int64); ok {
		r0 = rf(ctx, record, args...)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, types.Record, ...interface{}) error); ok {
		r1 = rf(ctx, record, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCollectionInterface_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type MockCollectionInterface_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - record types.Record
//   - args ...interface{}
func (_e *MockCollectionInterface_Expecter) Update(ctx interface{}, record interface{}, args ...interface{}) *MockCollectionInterface_Update_Call {
	return &MockCollectionInterface_Update_Call{Call: _e.mock.On("Update",
		append([]interface{}{ctx, record}, args...)...)}
}

func (_c *MockCollectionInterface_Update_Call) Run(run func(ctx context.Context, record types.Record, args ...interface{})) *MockCollectionInterface_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(context.Context), args[1].(types.Record), variadicArgs...)
	})
	return _c
}

func (_c *MockCollectionInterface_Update_Call) Return(_a0 int64, _a1 error) *MockCollectionInterface_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCollectionInterface_Update_Call) RunAndReturn(run func(context.Context, types.Record, ...interface{}) (int64, error)) *MockCollectionInterface_Update_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateRecord provides a mock function with given fields: ctx, id, record
func (_m *MockCollectionInterface) UpdateRecord(ctx context.Context, id interface{}, record types.Record) (*types.Record, error) {
	ret := _m.Called(ctx, id, record)

	if len(ret) == 0 {
		panic("no return value specified for UpdateRecord")
	}

	var r0 *types.Record
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, types.Record) (*types.Record, error)); ok {
		return rf(ctx, id, record)
	}
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, types.Record) *types.Record); ok {
		r0 = rf(ctx, id, record)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Record)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, interface{}, types.Record) error); ok {
		r1 = rf(ctx, id, record)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCollectionInterface_UpdateRecord_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateRecord'
type MockCollectionInterface_UpdateRecord_Call struct {
	*mock.Call
}

// UpdateRecord is a helper method to define mock.On call
//   - ctx context.Context
//   - id interface{}
//   - record types.Record
func (_e *MockCollectionInterface_Expecter) UpdateRecord(ctx interface{}, id interface{}, record interface{}) *MockCollectionInterface_UpdateRecord_Call {
	return &MockCollectionInterface_UpdateRecord_Call{Call: _e.mock.On("UpdateRecord", ctx, id, record)}
}

func (_c *MockCollectionInterface_UpdateRecord_Call) Run(run func(ctx context.Context, id interface{}, record types.Record)) *MockCollectionInterface_UpdateRecord_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interface{}), args[2].(types.Record))
	})
	return _c
}

func (_c *MockCollectionInterface_UpdateRecord_Call) Return(_a0 *types.Record, _a1 error) *MockCollectionInterface_UpdateRecord_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCollectionInterface_UpdateRecord_Call) RunAndReturn(run func(context.Context, interface{}, types.Record) (*types.Record, error)) *MockCollectionInterface_UpdateRecord_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockCollectionInterface creates a new instance of MockCollectionInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCollectionInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCollectionInterface {
	mock := &MockCollectionInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
