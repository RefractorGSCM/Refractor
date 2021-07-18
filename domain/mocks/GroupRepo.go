// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	domain "Refractor/domain"
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// GroupRepo is an autogenerated mock type for the GroupRepo type
type GroupRepo struct {
	mock.Mock
}

// Delete provides a mock function with given fields: ctx, id
func (_m *GroupRepo) Delete(ctx context.Context, id int64) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAll provides a mock function with given fields: ctx
func (_m *GroupRepo) GetAll(ctx context.Context) ([]*domain.Group, error) {
	ret := _m.Called(ctx)

	var r0 []*domain.Group
	if rf, ok := ret.Get(0).(func(context.Context) []*domain.Group); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Group)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBaseGroup provides a mock function with given fields: ctx
func (_m *GroupRepo) GetBaseGroup(ctx context.Context) (*domain.Group, error) {
	ret := _m.Called(ctx)

	var r0 *domain.Group
	if rf, ok := ret.Get(0).(func(context.Context) *domain.Group); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Group)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *GroupRepo) GetByID(ctx context.Context, id int64) (*domain.Group, error) {
	ret := _m.Called(ctx, id)

	var r0 *domain.Group
	if rf, ok := ret.Get(0).(func(context.Context, int64) *domain.Group); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Group)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserGroups provides a mock function with given fields: ctx, userID
func (_m *GroupRepo) GetUserGroups(ctx context.Context, userID string) ([]*domain.Group, error) {
	ret := _m.Called(ctx, userID)

	var r0 []*domain.Group
	if rf, ok := ret.Get(0).(func(context.Context, string) []*domain.Group); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Group)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserOverrides provides a mock function with given fields: ctx, userID
func (_m *GroupRepo) GetUserOverrides(ctx context.Context, userID string) (*domain.Overrides, error) {
	ret := _m.Called(ctx, userID)

	var r0 *domain.Overrides
	if rf, ok := ret.Get(0).(func(context.Context, string) *domain.Overrides); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Overrides)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetBaseGroup provides a mock function with given fields: ctx, group
func (_m *GroupRepo) SetBaseGroup(ctx context.Context, group *domain.Group) error {
	ret := _m.Called(ctx, group)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Group) error); ok {
		r0 = rf(ctx, group)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetUserOverrides provides a mock function with given fields: ctx, userID, overrides
func (_m *GroupRepo) SetUserOverrides(ctx context.Context, userID string, overrides *domain.Overrides) error {
	ret := _m.Called(ctx, userID, overrides)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *domain.Overrides) error); ok {
		r0 = rf(ctx, userID, overrides)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Store provides a mock function with given fields: ctx, group
func (_m *GroupRepo) Store(ctx context.Context, group *domain.Group) error {
	ret := _m.Called(ctx, group)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Group) error); ok {
		r0 = rf(ctx, group)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
