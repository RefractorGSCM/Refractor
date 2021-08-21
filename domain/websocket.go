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

package domain

import (
	"Refractor/pkg/broadcast"
	"net"
)

type WebsocketMessage struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
}

type WebsocketDirectMessage struct {
	ClientID int64
	Message  *WebsocketMessage
}

type WebsocketService interface {
	CreateClient(userID string, conn net.Conn)
	StartPool()
	Broadcast(message *WebsocketMessage)
	HandlePlayerJoin(fields broadcast.Fields, serverID int64, gameConfig *GameConfig)
	HandlePlayerQuit(fields broadcast.Fields, serverID int64, gameConfig *GameConfig)
}
