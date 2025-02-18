// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// RefreshTokenModifier is an autogenerated mock type for the RefreshTokenModifier type
type RefreshTokenModifier struct {
	mock.Mock
}

// ReplaceRefreshToken provides a mock function with given fields: ctx, oldID, newID, userID, expiry
func (_m *RefreshTokenModifier) ReplaceRefreshToken(ctx context.Context, oldID string, newID string, userID string, expiry time.Duration) error {
	ret := _m.Called(ctx, oldID, newID, userID, expiry)

	if len(ret) == 0 {
		panic("no return value specified for ReplaceRefreshToken")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, time.Duration) error); ok {
		r0 = rf(ctx, oldID, newID, userID, expiry)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetRefreshToken provides a mock function with given fields: ctx, userID, tokenID, expiry
func (_m *RefreshTokenModifier) SetRefreshToken(ctx context.Context, userID string, tokenID string, expiry time.Duration) error {
	ret := _m.Called(ctx, userID, tokenID, expiry)

	if len(ret) == 0 {
		panic("no return value specified for SetRefreshToken")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, time.Duration) error); ok {
		r0 = rf(ctx, userID, tokenID, expiry)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewRefreshTokenModifier creates a new instance of RefreshTokenModifier. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRefreshTokenModifier(t interface {
	mock.TestingT
	Cleanup(func())
}) *RefreshTokenModifier {
	mock := &RefreshTokenModifier{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
