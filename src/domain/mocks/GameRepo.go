// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	domain "Refractor/domain"

	mock "github.com/stretchr/testify/mock"
)

// GameRepo is an autogenerated mock type for the GameRepo type
type GameRepo struct {
	mock.Mock
}

// GetSettings provides a mock function with given fields: game
func (_m *GameRepo) GetSettings(game domain.Game) (*domain.GameSettings, error) {
	ret := _m.Called(game)

	var r0 *domain.GameSettings
	if rf, ok := ret.Get(0).(func(domain.Game) *domain.GameSettings); ok {
		r0 = rf(game)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.GameSettings)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(domain.Game) error); ok {
		r1 = rf(game)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetSettings provides a mock function with given fields: game, settings
func (_m *GameRepo) SetSettings(game domain.Game, settings *domain.GameSettings) error {
	ret := _m.Called(game, settings)

	var r0 error
	if rf, ok := ret.Get(0).(func(domain.Game, *domain.GameSettings) error); ok {
		r0 = rf(game, settings)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
