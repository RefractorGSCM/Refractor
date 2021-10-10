// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	bitperms "Refractor/pkg/bitperms"
	context "context"

	domain "Refractor/domain"

	mock "github.com/stretchr/testify/mock"
)

// Authorizer is an autogenerated mock type for the Authorizer type
type Authorizer struct {
	mock.Mock
}

// GetPermissions provides a mock function with given fields: ctx, scope, userID
func (_m *Authorizer) GetPermissions(ctx context.Context, scope domain.AuthScope, userID string) (*bitperms.Permissions, error) {
	ret := _m.Called(ctx, scope, userID)

	var r0 *bitperms.Permissions
	if rf, ok := ret.Get(0).(func(context.Context, domain.AuthScope, string) *bitperms.Permissions); ok {
		r0 = rf(ctx, scope, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*bitperms.Permissions)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, domain.AuthScope, string) error); ok {
		r1 = rf(ctx, scope, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HasPermission provides a mock function with given fields: ctx, scope, userID, authChecker
func (_m *Authorizer) HasPermission(ctx context.Context, scope domain.AuthScope, userID string, authChecker domain.AuthChecker) (bool, error) {
	ret := _m.Called(ctx, scope, userID, authChecker)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, domain.AuthScope, string, domain.AuthChecker) bool); ok {
		r0 = rf(ctx, scope, userID, authChecker)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, domain.AuthScope, string, domain.AuthChecker) error); ok {
		r1 = rf(ctx, scope, userID, authChecker)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}