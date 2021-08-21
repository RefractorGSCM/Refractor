/*
 * This file is part of Refractor.
 *
 * Refractor is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package service

import (
	"Refractor/domain"
	"Refractor/pkg/broadcast"
	"Refractor/pkg/websocket"
	"go.uber.org/zap"
	"net"
)

type websocketService struct {
	pool   *websocket.Pool
	logger *zap.Logger
}

func NewWebsocketService(log *zap.Logger) domain.WebsocketService {
	return &websocketService{
		pool:   websocket.NewPool(log),
		logger: log,
	}
}

func (s *websocketService) CreateClient(userID string, conn net.Conn) {
	client := websocket.NewClient(userID, conn, s.pool, s.logger)

	s.pool.Register <- client
	client.Read()
}

func (s *websocketService) StartPool() {
	s.pool.Start()
}

func (s *websocketService) Broadcast(message *domain.WebsocketMessage) {
	s.pool.Broadcast <- message
}

func (s *websocketService) HandlePlayerJoin(fields broadcast.Fields, serverID int64, game domain.Game) {
	panic("implement me")
}

func (s *websocketService) HandlePlayerQuit(fields broadcast.Fields, serverID int64, game domain.Game) {
	panic("implement me")
}
