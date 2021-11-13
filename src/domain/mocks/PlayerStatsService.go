// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// PlayerStatsService is an autogenerated mock type for the PlayerStatsService type
type PlayerStatsService struct {
	mock.Mock
}

// GetInfractionCount provides a mock function with given fields: c, platform, playerID
func (_m *PlayerStatsService) GetInfractionCount(c context.Context, platform string, playerID string) (int, error) {
	ret := _m.Called(c, platform, playerID)

	var r0 int
	if rf, ok := ret.Get(0).(func(context.Context, string, string) int); ok {
		r0 = rf(c, platform, playerID)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(c, platform, playerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetInfractionCountSince provides a mock function with given fields: c, platform, playerID, sinceMinutes
func (_m *PlayerStatsService) GetInfractionCountSince(c context.Context, platform string, playerID string, sinceMinutes int) (int, error) {
	ret := _m.Called(c, platform, playerID, sinceMinutes)

	var r0 int
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) int); ok {
		r0 = rf(c, platform, playerID, sinceMinutes)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, int) error); ok {
		r1 = rf(c, platform, playerID, sinceMinutes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
