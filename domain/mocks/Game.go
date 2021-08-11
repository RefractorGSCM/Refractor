// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	domain "Refractor/domain"

	mock "github.com/stretchr/testify/mock"
)

// Game is an autogenerated mock type for the Game type
type Game struct {
	mock.Mock
}

// GetCommandOutputPatterns provides a mock function with given fields:
func (_m *Game) GetCommandOutputPatterns() *domain.CommandOutputPatterns {
	ret := _m.Called()

	var r0 *domain.CommandOutputPatterns
	if rf, ok := ret.Get(0).(func() *domain.CommandOutputPatterns); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.CommandOutputPatterns)
		}
	}

	return r0
}

// GetConfig provides a mock function with given fields:
func (_m *Game) GetConfig() *domain.GameConfig {
	ret := _m.Called()

	var r0 *domain.GameConfig
	if rf, ok := ret.Get(0).(func() *domain.GameConfig); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.GameConfig)
		}
	}

	return r0
}

// GetName provides a mock function with given fields:
func (_m *Game) GetName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetPlatform provides a mock function with given fields:
func (_m *Game) GetPlatform() domain.Platform {
	ret := _m.Called()

	var r0 domain.Platform
	if rf, ok := ret.Get(0).(func() domain.Platform); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(domain.Platform)
		}
	}

	return r0
}

// GetPlayerListCommand provides a mock function with given fields:
func (_m *Game) GetPlayerListCommand() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
