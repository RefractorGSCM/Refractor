// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	broadcast "Refractor/pkg/broadcast"
	context "context"

	domain "Refractor/domain"

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

// HandleModerationAction provides a mock function with given fields: fields, serverID, game
func (_m *InfractionService) HandleModerationAction(fields broadcast.Fields, serverID int64, game domain.Game) {
	_m.Called(fields, serverID, game)
}

// HandlePlayerJoin provides a mock function with given fields: fields, serverID, game
func (_m *InfractionService) HandlePlayerJoin(fields broadcast.Fields, serverID int64, game domain.Game) {
	_m.Called(fields, serverID, game)
}

// LinkChatMessages provides a mock function with given fields: c, id, messageIDs
func (_m *InfractionService) LinkChatMessages(c context.Context, id int64, messageIDs ...int64) error {
	_va := make([]interface{}, len(messageIDs))
	for _i := range messageIDs {
		_va[_i] = messageIDs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, c, id)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, ...int64) error); ok {
		r0 = rf(c, id, messageIDs...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PlayerIsBanned provides a mock function with given fields: c, platform, playerID
func (_m *InfractionService) PlayerIsBanned(c context.Context, platform string, playerID string) (bool, int64, error) {
	ret := _m.Called(c, platform, playerID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(c, platform, playerID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 int64
	if rf, ok := ret.Get(1).(func(context.Context, string, string) int64); ok {
		r1 = rf(c, platform, playerID)
	} else {
		r1 = ret.Get(1).(int64)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string, string) error); ok {
		r2 = rf(c, platform, playerID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// PlayerIsMuted provides a mock function with given fields: c, platform, playerID
func (_m *InfractionService) PlayerIsMuted(c context.Context, platform string, playerID string) (bool, int64, error) {
	ret := _m.Called(c, platform, playerID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(c, platform, playerID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 int64
	if rf, ok := ret.Get(1).(func(context.Context, string, string) int64); ok {
		r1 = rf(c, platform, playerID)
	} else {
		r1 = ret.Get(1).(int64)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string, string) error); ok {
		r2 = rf(c, platform, playerID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// SetRepealed provides a mock function with given fields: c, id, repealed
func (_m *InfractionService) SetRepealed(c context.Context, id int64, repealed bool) (*domain.Infraction, error) {
	ret := _m.Called(c, id, repealed)

	var r0 *domain.Infraction
	if rf, ok := ret.Get(0).(func(context.Context, int64, bool) *domain.Infraction); ok {
		r0 = rf(c, id, repealed)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Infraction)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64, bool) error); ok {
		r1 = rf(c, id, repealed)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store provides a mock function with given fields: c, infraction, attachments, linkedMessages
func (_m *InfractionService) Store(c context.Context, infraction *domain.Infraction, attachments []*domain.Attachment, linkedMessages []int64) (*domain.Infraction, error) {
	ret := _m.Called(c, infraction, attachments, linkedMessages)

	var r0 *domain.Infraction
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Infraction, []*domain.Attachment, []int64) *domain.Infraction); ok {
		r0 = rf(c, infraction, attachments, linkedMessages)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Infraction)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *domain.Infraction, []*domain.Attachment, []int64) error); ok {
		r1 = rf(c, infraction, attachments, linkedMessages)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SubscribeInfractionCreate provides a mock function with given fields: sub
func (_m *InfractionService) SubscribeInfractionCreate(sub domain.InfractionSubscriber) {
	_m.Called(sub)
}

// UnlinkChatMessages provides a mock function with given fields: c, id, messageIDs
func (_m *InfractionService) UnlinkChatMessages(c context.Context, id int64, messageIDs ...int64) error {
	_va := make([]interface{}, len(messageIDs))
	for _i := range messageIDs {
		_va[_i] = messageIDs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, c, id)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, ...int64) error); ok {
		r0 = rf(c, id, messageIDs...)
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
