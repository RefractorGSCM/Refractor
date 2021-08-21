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
)

type ClientCreator interface {
	GetClientFromConfig(gameConfig *GameConfig, server *Server) (RCONClient, error)
}

type RCONClient interface {
	ExecCommand(string) (string, error)
	Connect() error
	ListenForBroadcasts([]string, chan error)
	SetBroadcastHandler(handlerFunc rcon.BroadcastHandlerFunc)
	SetDisconnectHandler(handlerFunc rcon.DisconnectHandlerFunc)
}

type BroadcastSubscriber func(fields broadcast.Fields, serverID int64, gameConfig *GameConfig)

type RCONService interface {
	CreateClient(server *Server) error
	GetClients() map[int64]RCONClient
	DeleteClient(serverID int64)
	GetServerClient(serverID int64) RCONClient
	StartReconnectRoutine(server *Server, data *ServerData)
	SubscribeJoin(sub BroadcastSubscriber)
	SubscribeQuit(sub BroadcastSubscriber)
}
