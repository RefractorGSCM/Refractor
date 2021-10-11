// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	domain "Refractor/domain"
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// InfractionService is an autogenerated mock type for the InfractionService type
type InfractionService struct {
	mock.Mock
}

// Delete provides a mock function with given fields: c, id
func (_m *InfractionService) Delete(c context.Context, id int64) error {
	ret := _m.Called(c, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(c, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByID provides a mock function with given fields: c, id
func (_m *InfractionService) GetByID(c context.Context, id int64) (*domain.Infraction, error) {
	ret := _m.Called(c, id)

	var r0 *domain.Infraction
	if rf, ok := ret.Get(0).(func(context.Context, int64) *domain.Infraction); ok {
		r0 = rf(c, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Infraction)
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

// GetByPlayer provides a mock function with given fields: c, playerID, platform
func (_m *InfractionService) GetByPlayer(c context.Context, playerID string, platform string) ([]*domain.Infraction, error) {
	ret := _m.Called(c, playerID, platform)

	var r0 []*domain.Infraction
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []*domain.Infraction); ok {
		r0 = rf(c, playerID, platform)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.Infraction)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(c, playerID, platform)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLinkedChatMessages provides a mock function with given fields: c, id
func (_m *InfractionService) GetLinkedChatMessages(c context.Context, id int64) ([]*domain.ChatMessage, error) {
	ret := _m.Called(c, id)

	var r0 []*domain.ChatMessage
	if rf, ok := ret.Get(0).(func(context.Context, int64) []*domain.ChatMessage); ok {
		r0 = rf(c, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.ChatMessage)
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

// LinkChatMessage provides a mock function with given fields: c, id, messageID
func (_m *InfractionService) LinkChatMessage(c context.Context, id int64, messageID int64) error {
	ret := _m.Called(c, id, messageID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) error); ok {
		r0 = rf(c, id, messageID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Store provides a mock function with given fields: c, infraction, attachments
func (_m *InfractionService) Store(c context.Context, infraction *domain.Infraction, attachments []*domain.Attachment) (*domain.Infraction, error) {
	ret := _m.Called(c, infraction, attachments)

	var r0 *domain.Infraction
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Infraction, []*domain.Attachment) *domain.Infraction); ok {
		r0 = rf(c, infraction, attachments)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Infraction)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *domain.Infraction, []*domain.Attachment) error); ok {
		r1 = rf(c, infraction, attachments)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnlinkChatMessage provides a mock function with given fields: c, id, messageID
func (_m *InfractionService) UnlinkChatMessage(c context.Context, id int64, messageID int64) error {
	ret := _m.Called(c, id, messageID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) error); ok {
		r0 = rf(c, id, messageID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: c, id, args
func (_m *InfractionService) Update(c context.Context, id int64, args domain.UpdateArgs) (*domain.Infraction, error) {
	ret := _m.Called(c, id, args)

	var r0 *domain.Infraction
	if rf, ok := ret.Get(0).(func(context.Context, int64, domain.UpdateArgs) *domain.Infraction); ok {
		r0 = rf(c, id, args)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Infraction)
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
