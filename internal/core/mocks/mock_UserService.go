// Code generated by mockery v2.43.2. DO NOT EDIT.

package core

import (
	context "context"

	core "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	mock "github.com/stretchr/testify/mock"
)

// MockUserService is an autogenerated mock type for the UserService type
type MockUserService struct {
	mock.Mock
}

type MockUserService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUserService) EXPECT() *MockUserService_Expecter {
	return &MockUserService_Expecter{mock: &_m.Mock}
}

// DeleteUser provides a mock function with given fields: ctx, userID
func (_m *MockUserService) DeleteUser(ctx context.Context, userID int) error {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for DeleteUser")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockUserService_DeleteUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteUser'
type MockUserService_DeleteUser_Call struct {
	*mock.Call
}

// DeleteUser is a helper method to define mock.On call
//   - ctx context.Context
//   - userID int
func (_e *MockUserService_Expecter) DeleteUser(ctx interface{}, userID interface{}) *MockUserService_DeleteUser_Call {
	return &MockUserService_DeleteUser_Call{Call: _e.mock.On("DeleteUser", ctx, userID)}
}

func (_c *MockUserService_DeleteUser_Call) Run(run func(ctx context.Context, userID int)) *MockUserService_DeleteUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int))
	})
	return _c
}

func (_c *MockUserService_DeleteUser_Call) Return(_a0 error) *MockUserService_DeleteUser_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockUserService_DeleteUser_Call) RunAndReturn(run func(context.Context, int) error) *MockUserService_DeleteUser_Call {
	_c.Call.Return(run)
	return _c
}

// GetUser provides a mock function with given fields: ctx, user
func (_m *MockUserService) GetUser(ctx context.Context, user core.User) (*core.User, error) {
	ret := _m.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for GetUser")
	}

	var r0 *core.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, core.User) (*core.User, error)); ok {
		return rf(ctx, user)
	}
	if rf, ok := ret.Get(0).(func(context.Context, core.User) *core.User); ok {
		r0 = rf(ctx, user)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, core.User) error); ok {
		r1 = rf(ctx, user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserService_GetUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUser'
type MockUserService_GetUser_Call struct {
	*mock.Call
}

// GetUser is a helper method to define mock.On call
//   - ctx context.Context
//   - user core.User
func (_e *MockUserService_Expecter) GetUser(ctx interface{}, user interface{}) *MockUserService_GetUser_Call {
	return &MockUserService_GetUser_Call{Call: _e.mock.On("GetUser", ctx, user)}
}

func (_c *MockUserService_GetUser_Call) Run(run func(ctx context.Context, user core.User)) *MockUserService_GetUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(core.User))
	})
	return _c
}

func (_c *MockUserService_GetUser_Call) Return(_a0 *core.User, _a1 error) *MockUserService_GetUser_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserService_GetUser_Call) RunAndReturn(run func(context.Context, core.User) (*core.User, error)) *MockUserService_GetUser_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateUser provides a mock function with given fields: ctx, user
func (_m *MockUserService) UpdateUser(ctx context.Context, user core.UpdateUser) (*core.User, error) {
	ret := _m.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for UpdateUser")
	}

	var r0 *core.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, core.UpdateUser) (*core.User, error)); ok {
		return rf(ctx, user)
	}
	if rf, ok := ret.Get(0).(func(context.Context, core.UpdateUser) *core.User); ok {
		r0 = rf(ctx, user)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, core.UpdateUser) error); ok {
		r1 = rf(ctx, user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserService_UpdateUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateUser'
type MockUserService_UpdateUser_Call struct {
	*mock.Call
}

// UpdateUser is a helper method to define mock.On call
//   - ctx context.Context
//   - user core.UpdateUser
func (_e *MockUserService_Expecter) UpdateUser(ctx interface{}, user interface{}) *MockUserService_UpdateUser_Call {
	return &MockUserService_UpdateUser_Call{Call: _e.mock.On("UpdateUser", ctx, user)}
}

func (_c *MockUserService_UpdateUser_Call) Run(run func(ctx context.Context, user core.UpdateUser)) *MockUserService_UpdateUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(core.UpdateUser))
	})
	return _c
}

func (_c *MockUserService_UpdateUser_Call) Return(_a0 *core.User, _a1 error) *MockUserService_UpdateUser_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserService_UpdateUser_Call) RunAndReturn(run func(context.Context, core.UpdateUser) (*core.User, error)) *MockUserService_UpdateUser_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockUserService creates a new instance of MockUserService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUserService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserService {
	mock := &MockUserService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
