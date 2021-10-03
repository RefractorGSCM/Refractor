// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	domain "Refractor/domain"
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// FlaggedWordRepo is an autogenerated mock type for the FlaggedWordRepo type
type FlaggedWordRepo struct {
	mock.Mock
}

// Delete provides a mock function with given fields: ctx, id
func (_m *FlaggedWordRepo) Delete(ctx context.Context, id int64) error {
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
func (_m *FlaggedWordRepo) GetAll(ctx context.Context) ([]*domain.FlaggedWord, error) {
	ret := _m.Called(ctx)

	var r0 []*domain.FlaggedWord
	if rf, ok := ret.Get(0).(func(context.Context) []*domain.FlaggedWord); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*domain.FlaggedWord)
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

// Store provides a mock function with given fields: ctx, word
func (_m *FlaggedWordRepo) Store(ctx context.Context, word *domain.FlaggedWord) error {
	ret := _m.Called(ctx, word)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.FlaggedWord) error); ok {
		r0 = rf(ctx, word)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: ctx, id, newWord
func (_m *FlaggedWordRepo) Update(ctx context.Context, id int64, newWord string) (*domain.FlaggedWord, error) {
	ret := _m.Called(ctx, id, newWord)

	var r0 *domain.FlaggedWord
	if rf, ok := ret.Get(0).(func(context.Context, int64, string) *domain.FlaggedWord); ok {
		r0 = rf(ctx, id, newWord)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.FlaggedWord)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64, string) error); ok {
		r1 = rf(ctx, id, newWord)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
