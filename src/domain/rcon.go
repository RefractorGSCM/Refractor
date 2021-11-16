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
	"github.com/refractorgscm/rcon"
	"sync"
)

type ClientCreator interface {
	GetClientFromConfig(game Game, server *Server) (RCONClient, error)
}

type RCONClient interface {
	RunCommand(string) (string, error)
	Connect() error
	WaitGroup() *sync.WaitGroup
	SetBroadcastHandler(handlerFunc rcon.BroadcastHandler)
	SetDisconnectHandler(handlerFunc rcon.DisconnectHandler)
	SetBroadcastChecker(handlerFunc rcon.BroadcastMessageChecker)
	GetGame() Game
	Close() error
}

type OnlinePlayer struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
}

type BroadcastSubscriber func(fields broadcast.Fields, serverID int64, game Game)
type PlayerListUpdateSubscriber func(serverID int64, players []*OnlinePlayer, game Game)
type ServerStatusSubscriber func(serverID int64, status string)
type ChatReceiveSubscriber func(body *ChatReceiveBody, serverID int64, game Game)

type RCONService interface {
	CreateClient(server *Server) error
	GetClients() map[int64]RCONClient
	DeleteClient(serverID int64)
	GetServerClient(serverID int64) RCONClient
	RefreshPlayerList(serverID int64, game Game) error
	StartReconnectRoutine(serverID int64, data *ServerData)
	SubscribeJoin(sub BroadcastSubscriber)
	SubscribeQuit(sub BroadcastSubscriber)
	SubscribePlayerListUpdate(sub PlayerListUpdateSubscriber)
	SubscribeServerStatus(sub ServerStatusSubscriber)
	SubscribeChat(sub ChatReceiveSubscriber)
	SubscribeModeratorAction(sub BroadcastSubscriber)
	SendChatMessage(body *ChatSendBody)
	HandleServerUpdate(server *Server)
}
