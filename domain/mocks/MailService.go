// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// MailService is an autogenerated mock type for the MailService type
type MailService struct {
	mock.Mock
}

// SendMail provides a mock function with given fields: to, sub, body
func (_m *MailService) SendMail(to []string, sub string, body string) error {
	ret := _m.Called(to, sub, body)

	var r0 error
	if rf, ok := ret.Get(0).(func([]string, string, string) error); ok {
		r0 = rf(to, sub, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendWelcomeEmail provides a mock function with given fields: to, inviterName, link
func (_m *MailService) SendWelcomeEmail(to string, inviterName string, link string) error {
	ret := _m.Called(to, inviterName, link)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(to, inviterName, link)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
