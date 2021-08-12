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

package websocket

import (
	"Refractor/domain"
	"encoding/json"
	"github.com/gobwas/ws/wsutil"
	"go.uber.org/zap"
)

type Pool struct {
	Clients    map[int64]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *domain.WebsocketMessage
	SendDirect chan *domain.WebsocketDirectMessage
	logger     *zap.Logger
}

func NewPool(log *zap.Logger) *Pool {
	return &Pool{
		Clients:    map[int64]*Client{},
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *domain.WebsocketMessage),
		SendDirect: make(chan *domain.WebsocketDirectMessage),
		logger:     log,
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client.ID] = client

			pool.logger.Info(
				"Websocket client registered",
				zap.Int64("Client ID", client.ID),
				zap.String("User ID", client.UserID),
			)
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client.ID)

			pool.logger.Info(
				"Websocket client unregistered",
				zap.Int64("Client ID", client.ID),
				zap.String("User ID", client.UserID),
			)
			break
		case msg := <-pool.Broadcast:
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				pool.logger.Error("Could not marshal broadcast message", zap.Error(err))
				continue
			}

			for _, client := range pool.Clients {
				if err := wsutil.WriteServerText(client.Conn, msgBytes); err != nil {
					pool.logger.Error(
						"Could not send broadcast message to client",
						zap.Int64("Client ID", client.ID),
						zap.String("User ID", client.UserID),
						zap.Error(err),
					)
					continue
				}
			}
			break
		case sendParams := <-pool.SendDirect:
			msgBytes, err := json.Marshal(sendParams.Message)
			if err != nil {
				pool.logger.Error("Could not marshal broadcast message", zap.Error(err))
				continue
			}

			client := pool.Clients[sendParams.ClientID]
			if client == nil {
				pool.logger.Warn(
					"Tried to send direct message to non-existing client",
					zap.Int64("Client ID", client.ID),
				)
				continue
			}

			if err := wsutil.WriteServerText(client.Conn, msgBytes); err != nil {
				pool.logger.Error(
					"Could not send direct message to client",
					zap.Int64("Client ID", client.ID),
					zap.String("User ID", client.UserID),
					zap.Error(err),
				)
				continue
			}
			break
		}
	}
}
