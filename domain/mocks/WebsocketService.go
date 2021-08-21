// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	domain "Refractor/domain"
	broadcast "Refractor/pkg/broadcast"

	mock "github.com/stretchr/testify/mock"

	net "net"
)

// WebsocketService is an autogenerated mock type for the WebsocketService type
type WebsocketService struct {
	mock.Mock
}

// Broadcast provides a mock function with given fields: message
func (_m *WebsocketService) Broadcast(message *domain.WebsocketMessage) {
	_m.Called(message)
}

// CreateClient provides a mock function with given fields: userID, conn
func (_m *WebsocketService) CreateClient(userID string, conn net.Conn) {
	_m.Called(userID, conn)
}

// HandlePlayerJoin provides a mock function with given fields: fields, serverID, game
func (_m *WebsocketService) HandlePlayerJoin(fields broadcast.Fields, serverID int64, game domain.Game) {
	_m.Called(fields, serverID, game)
}

// HandlePlayerQuit provides a mock function with given fields: fields, serverID, game
func (_m *WebsocketService) HandlePlayerQuit(fields broadcast.Fields, serverID int64, game domain.Game) {
	_m.Called(fields, serverID, game)
}

// StartPool provides a mock function with given fields:
func (_m *WebsocketService) StartPool() {
	_m.Called()
}
