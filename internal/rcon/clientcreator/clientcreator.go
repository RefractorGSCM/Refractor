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

package clientcreator

import (
	"Refractor/domain"
	"github.com/refractorgscm/rcon"
	"strconv"
)

type clientCreator struct{}

type Client struct {
	game   domain.Game
	Server *domain.Server
	*rcon.Client
}

func (c *Client) GetGame() domain.Game {
	return c.game
}

func NewClientCreator() domain.ClientCreator {
	return &clientCreator{}
}

func (c *clientCreator) GetClientFromConfig(game domain.Game, server *domain.Server) (domain.RCONClient, error) {
	port, err := strconv.ParseUint(server.RCONPort, 10, 16)
	if err != nil {
		return nil, err
	}

	gameConfig := game.GetConfig()

	// Create RCON client
	client := rcon.NewClient(&rcon.ClientConfig{
		Host:                     server.Address,
		Port:                     uint16(port),
		Password:                 server.RCONPassword,
		SendHeartbeatCommand:     gameConfig.AlivePingEnabled(),
		AttemptReconnect:         false,
		HeartbeatCommandInterval: gameConfig.AlivePingInterval,
		EnableBroadcasts:         gameConfig.EnableBroadcasts,
		NonBroadcastPatterns:     gameConfig.IgnoredBroadcastPatterns,
	})

	return &Client{
		game:   game,
		Server: server,
		Client: client,
	}, nil
}
