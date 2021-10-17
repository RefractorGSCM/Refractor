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

type ChatSendBody struct {
	ServerID int64  `json:"server_id"`
	Message  string `json:"message"`
	Sender   string `json:"sender"`

	// SentByUser is true if this message was sent by another user
	SentByUser bool `json:"sent_by_user"`
}

type ChatSendSubscriber func(body *ChatSendBody)

type WebsocketService interface {
	CreateClient(userID string, conn net.Conn)
	StartPool()
	Broadcast(message *WebsocketMessage)
	BroadcastServerMessage(message *WebsocketMessage, serverID int64, authChecker AuthChecker) error
	SendDirectMessage(message *WebsocketMessage, userID string)
	HandlePlayerJoin(fields broadcast.Fields, serverID int64, game Game)
	HandlePlayerQuit(fields broadcast.Fields, serverID int64, game Game)
	HandleServerStatusChange(serverID int64, status string)
	SubscribeChatSend(sub ChatSendSubscriber)
}
