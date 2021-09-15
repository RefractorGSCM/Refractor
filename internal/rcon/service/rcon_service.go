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
	"Refractor/pkg/broadcast"
	"Refractor/pkg/regexutils"
	"fmt"
	"go.uber.org/zap"
	"net"
	"time"
)

type rconService struct {
	logger        *zap.Logger
	clients       map[int64]domain.RCONClient
	gameService   domain.GameService
	clientCreator domain.ClientCreator

	joinSubs       []domain.BroadcastSubscriber
	quitSubs       []domain.BroadcastSubscriber
	playerListSubs []domain.PlayerListUpdateSubscriber
	statusSubs     []domain.ServerStatusSubscriber
}

func NewRCONService(log *zap.Logger, gs domain.GameService) domain.RCONService {
	return &rconService{
		logger:         log,
		clients:        map[int64]domain.RCONClient{},
		gameService:    gs,
		clientCreator:  clientcreator.NewClientCreator(),
		joinSubs:       []domain.BroadcastSubscriber{},
		quitSubs:       []domain.BroadcastSubscriber{},
		playerListSubs: []domain.PlayerListUpdateSubscriber{},
		statusSubs:     []domain.ServerStatusSubscriber{},
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

	client.SetBroadcastHandler(s.getBroadcastHandler(server.ID, game))
	client.SetDisconnectHandler(s.getDisconnectHandler(server.ID))

	// Connect the main socket
	if err := client.Connect(); err != nil {
		return err
	}

	// Connect broadcast socket
	if gameConfig.EnableBroadcasts {
		errorChan := make(chan error)
		go client.ListenForBroadcasts(gameConfig.BroadcastInitCommands, errorChan)

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
	onlinePlayers, err := s.getOnlinePlayers(server.ID, game)
	if err != nil {
		return err
	}

	// Dispatch player join events for all currently online players
	for _, op := range onlinePlayers {
		fmt.Println(op.PlayerID, op.Name)
		for _, sub := range s.joinSubs {
			sub(broadcast.Fields{"PlayerID": op.PlayerID, "Name": op.Name}, server.ID, game)
		}
	}

	// Notify that this server is online
	for _, sub := range s.statusSubs {
		sub(server.ID, "Online")
	}

	return nil
}

func (s *rconService) GetClients() map[int64]domain.RCONClient {
	return s.clients
}

func (s *rconService) GetServerClient(serverID int64) domain.RCONClient {
	return s.clients[serverID]
}

func (s *rconService) DeleteClient(serverID int64) {
	delete(s.clients, serverID)
}

func (s *rconService) getBroadcastHandler(serverID int64, game domain.Game) func(string) {
	return func(message string) {
		s.logger.Info("Broadcast received", zap.Int64("Server", serverID), zap.String("Message", message))

		bcast := broadcast.GetBroadcastType(message, game.GetConfig().BroadcastPatterns)
		if bcast == nil {
			return
		}

		switch bcast.Type {
		case broadcast.TYPE_JOIN:
			for _, sub := range s.joinSubs {
				sub(bcast.Fields, serverID, game)
			}
			break
		case broadcast.TYPE_QUIT:
			for _, sub := range s.quitSubs {
				sub(bcast.Fields, serverID, game)
			}
			break
		}
	}
}

func (s *rconService) getDisconnectHandler(serverID int64) func(error, bool) {
	return func(err error, expected bool) {
		s.logger.Warn("RCON client disconnected", zap.Int64("Server", serverID), zap.Bool("Expected", expected), zap.Error(err))

		for _, sub := range s.statusSubs {
			sub(serverID, "Offline")
		}

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
		fmt.Println(player)

		fields := regexutils.MapNamedMatches(playerListPattern, player)

		onlinePlayers = append(onlinePlayers, &onlinePlayer{
			PlayerID: fields["PlayerID"],
			Name:     fields["Name"],
		})
	}

	return onlinePlayers, nil
}

func (s *rconService) StartReconnectRoutine(server *domain.Server, data *domain.ServerData) {
	var delay = time.Second * 5

	for {
		time.Sleep(delay)

		if err := s.CreateClient(server); err != nil {
			switch errType := err.(type) {
			case *net.OpError:
				// If this error is a dial error, we don't log it. If it isn't, we do want to log it.
				// We disregard dial errors because we can assume this means that the server is offline (in most cases).
				if errType.Op != "dial" {
					s.logger.Warn(
						"An RCON reconnect routine connection error has occurred",
						zap.Int64("Server", server.ID),
						zap.Error(err),
					)
				}
				break
			default:
				s.logger.Error(
					"RCON reconnect routine could not create a new client for server",
					zap.Int64("Server", server.ID),
					zap.Error(err),
				)
				continue
			}
		} else {
			s.logger.Info(
				"RCON connection established to server",
				zap.Int64("Server", server.ID),
			)

			data.ReconnectInProgress = false
			break
		}

		if delay < time.Minute*2 {
			delay += delay / 2
		} else {
			delay = time.Minute * 2
		}

		delay = delay.Round(time.Second)
		s.logger.Info(
			"Could not establish connection to server. Retrying later.",
			zap.Int64("Server", server.ID),
			zap.Duration("Retrying In", delay),
		)
	}

	s.logger.Info("Reconnect routine terminated", zap.Int64("Server", server.ID))
}

func (s *rconService) SubscribeJoin(sub domain.BroadcastSubscriber) {
	s.joinSubs = append(s.joinSubs, sub)
}

func (s *rconService) SubscribeQuit(sub domain.BroadcastSubscriber) {
	s.quitSubs = append(s.quitSubs, sub)
}

func (s *rconService) SubscribePlayerListUpdate(sub domain.PlayerListUpdateSubscriber) {
	s.playerListSubs = append(s.playerListSubs, sub)
}

func (s *rconService) SubscribeServerStatus(sub domain.ServerStatusSubscriber) {
	s.statusSubs = append(s.statusSubs, sub)
}
