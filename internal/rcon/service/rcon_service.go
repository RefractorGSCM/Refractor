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
	"Refractor/internal/rcon/clientcreator"
	"Refractor/pkg/regexutils"
	"go.uber.org/zap"
)

type rconService struct {
	logger        *zap.Logger
	clients       map[int64]domain.RCONClient
	gameService   domain.GameService
	clientCreator domain.ClientCreator
}

func NewRCONService(log *zap.Logger, gs domain.GameService) domain.RCONService {
	return &rconService{
		logger:        log,
		clients:       map[int64]domain.RCONClient{},
		gameService:   gs,
		clientCreator: clientcreator.NewClientCreator(),
	}
}

func (s *rconService) CreateClient(server *domain.Server) error {
	// Get the server's game
	game, err := s.gameService.GetGame(server.Game)
	if err != nil {
		return err
	}

	gameConfig := game.GetConfig()

	client, err := s.clientCreator.GetClientFromConfig(gameConfig, server)
	if err != nil {
		return err
	}

	client.SetBroadcastHandler(s.getBroadcastHandler(server.ID, gameConfig))
	client.SetDisconnectHandler(s.getDisconnectHandler(server.ID))

	// Connect the main socket
	if err := client.Connect(); err != nil {
		return err
	}

	// Connect broadcast socket
	if gameConfig.EnableBroadcasts {
		errorChan := make(chan error)
		go client.ListenForBroadcasts([]string{"login, chat"}, errorChan)

		go func() {
			select {
			case err := <-errorChan:
				s.logger.Error("Broadcast listener error", zap.Int64("Server", server.ID), zap.Error(err))
				break
			}
		}()
	}

	if gameConfig.PlayerListPollingEnabled() {
		// TODO: Start player list polling routing
	}

	// Add to list of clients
	s.clients[server.ID] = client

	// Get currently online players
	//onlinePlayers := s.getOnlinePlayers(server.ID, game)

	return nil
}

func (s *rconService) GetClients() map[int64]domain.RCONClient {
	panic("implement me")
}

func (s *rconService) GetServerClient(serverID int64) domain.RCONClient {
	return s.clients[serverID]
}

func (s *rconService) DeleteClient(serverID int64) {
	delete(s.clients, serverID)
}

func (s *rconService) getBroadcastHandler(serverID int64, gameConfig *domain.GameConfig) func(string) {
	return func(message string) {
		s.logger.Info("Broadcast received", zap.Int64("Server", serverID), zap.String("Message", message))
	}
}

func (s *rconService) getDisconnectHandler(serverID int64) func(error, bool) {
	return func(err error, expected bool) {
		s.logger.Warn("RCON client disconnected", zap.Int64("Server", serverID), zap.Bool("Expected", expected), zap.Error(err))

		// Delete the client from the list of clients. Reconnection attempts will be made in the watchdog.
		s.DeleteClient(serverID)
	}
}

type onlinePlayer struct {
	PlayerID string
	Name     string
}

func (s *rconService) getOnlinePlayers(serverID int64, game domain.Game) ([]*onlinePlayer, error) {
	playerListCommand := game.GetPlayerListCommand()

	res, err := s.GetServerClient(serverID).ExecCommand(playerListCommand)
	if err != nil {
		s.logger.Error(
			"Could not execute RCON player list command",
			zap.Int64("Server", serverID),
			zap.String("Command", playerListCommand),
			zap.Error(err),
		)
		return nil, err
	}

	// Extract player info from RCON command response
	playerListPattern := game.GetCommandOutputPatterns().PlayerList
	players := playerListPattern.FindAllString(res, -1)

	var onlinePlayers []*onlinePlayer

	for _, player := range players {
		fields := regexutils.MapNamedMatches(playerListPattern, player)

		onlinePlayers = append(onlinePlayers, &onlinePlayer{
			PlayerID: fields["PlayerID"],
			Name:     fields["Name"],
		})
	}

	return onlinePlayers, nil
}
