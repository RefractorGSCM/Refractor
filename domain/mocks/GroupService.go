// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	domain "Refractor/domain"
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// GroupService is an autogenerated mock type for the GroupService type
type GroupService struct {
	mock.Mock
}

// AddUserGroup provides a mock function with given fields: c, groupctx
func (_m *GroupService) AddUserGroup(c context.Context, groupctx domain.GroupSetContext) error {
	ret := _m.Called(c, groupctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, domain.GroupSetContext) error); ok {
		r0 = rf(c, groupctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: c, id
func (_m *GroupService) Delete(c context.Context, id int64) error {
	ret := _m.Called(c, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(c, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAll provides a mock function with given fields: c
func (_m *GroupService) GetAll(c context.Context) ([]*domain.Group, error) {
	ret := _m.Called(c)

	var r0 []*domain.Group
	if rf, ok := ret.Get(0).(func(context.Context) []*domain.Group); ok {
		r0 = rf(c)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Group)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(c)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: c, id
func (_m *GroupService) GetByID(c context.Context, id int64) (*domain.Group, error) {
	ret := _m.Called(c, id)

	var r0 *domain.Group
	if rf, ok := ret.Get(0).(func(context.Context, int64) *domain.Group); ok {
		r0 = rf(c, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Group)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(c, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetServerOverridesAllGroups provides a mock function with given fields: c, serverID
func (_m *GroupService) GetServerOverridesAllGroups(c context.Context, serverID int64) ([]*domain.Overrides, error) {
	ret := _m.Called(c, serverID)

	var r0 []*domain.Overrides
	if rf, ok := ret.Get(0).(func(context.Context, int64) []*domain.Overrides); ok {
		r0 = rf(c, serverID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Overrides)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(c, serverID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveUserGroup provides a mock function with given fields: c, groupctx
func (_m *GroupService) RemoveUserGroup(c context.Context, groupctx domain.GroupSetContext) error {
	ret := _m.Called(c, groupctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, domain.GroupSetContext) error); ok {
		r0 = rf(c, groupctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Reorder provides a mock function with given fields: c, newPositions
func (_m *GroupService) Reorder(c context.Context, newPositions []*domain.GroupReorderInfo) error {
	ret := _m.Called(c, newPositions)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*domain.GroupReorderInfo) error); ok {
		r0 = rf(c, newPositions)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetServerOverrides provides a mock function with given fields: c, serverID, groupID, overrides
func (_m *GroupService) SetServerOverrides(c context.Context, serverID int64, groupID int64, overrides *domain.Overrides) error {
	ret := _m.Called(c, serverID, groupID, overrides)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64, *domain.Overrides) error); ok {
		r0 = rf(c, serverID, groupID, overrides)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Store provides a mock function with given fields: c, group
func (_m *GroupService) Store(c context.Context, group *domain.Group) error {
	ret := _m.Called(c, group)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Group) error); ok {
		r0 = rf(c, group)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: c, id, args
func (_m *GroupService) Update(c context.Context, id int64, args domain.UpdateArgs) (*domain.Group, error) {
	ret := _m.Called(c, id, args)

	var r0 *domain.Group
	if rf, ok := ret.Get(0).(func(context.Context, int64, domain.UpdateArgs) *domain.Group); ok {
		r0 = rf(c, id, args)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Group)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64, domain.UpdateArgs) error); ok {
		r1 = rf(c, id, args)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateBase provides a mock function with given fields: c, args
func (_m *GroupService) UpdateBase(c context.Context, args domain.UpdateArgs) (*domain.Group, error) {
	ret := _m.Called(c, args)

	var r0 *domain.Group
	if rf, ok := ret.Get(0).(func(context.Context, domain.UpdateArgs) *domain.Group); ok {
		r0 = rf(c, args)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Group)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, domain.UpdateArgs) error); ok {
		r1 = rf(c, args)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
