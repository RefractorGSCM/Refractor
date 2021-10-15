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
	"context"
	"fmt"
	"go.uber.org/zap"
	"net"
	"strconv"
	"strings"
	"time"
)

type rconService struct {
	logger            *zap.Logger
	clients           map[int64]domain.RCONClient
	gameService       domain.GameService
	infractionService domain.InfractionService
	clientCreator     domain.ClientCreator

	joinSubs       []domain.BroadcastSubscriber
	quitSubs       []domain.BroadcastSubscriber
	playerListSubs []domain.PlayerListUpdateSubscriber
	statusSubs     []domain.ServerStatusSubscriber
	chatSubs       []domain.ChatReceiveSubscriber
	prevPlayers    map[int64]map[string]*onlinePlayer
}

func NewRCONService(log *zap.Logger, gs domain.GameService, is domain.InfractionService) domain.RCONService {
	return &rconService{
		logger:            log,
		clients:           map[int64]domain.RCONClient{},
		gameService:       gs,
		infractionService: is,
		clientCreator:     clientcreator.NewClientCreator(),
		joinSubs:          []domain.BroadcastSubscriber{},
		quitSubs:          []domain.BroadcastSubscriber{},
		playerListSubs:    []domain.PlayerListUpdateSubscriber{},
		statusSubs:        []domain.ServerStatusSubscriber{},
		chatSubs:          []domain.ChatReceiveSubscriber{},
		prevPlayers:       map[int64]map[string]*onlinePlayer{},
	}
}

func (s *rconService) CreateClient(server *domain.Server) error {
	if !s.gameService.GameExists(server.Game) {
		return fmt.Errorf("could not create RCON client for servers with a non-existent game: %s", server.Game)
	}

	// Check if a client already exists. If one does, close the associated connections and delete it.
	if s.clients[server.ID] != nil {
		_ = s.clients[server.ID].Disconnect()
		delete(s.clients, server.ID)
	}

	// Get the server's game
	game, err := s.gameService.GetGame(server.Game)
	if err != nil {
		return err
	}

	gameConfig := game.GetConfig()

	client, err := s.clientCreator.GetClientFromConfig(game, server)
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
		go s.startPlayerListPolling(server.ID, game)
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
		s.HandlePlayerJoin(broadcast.Fields{
			"PlayerID": op.PlayerID,
			"Platform": game.GetPlatform().GetName(),
			"Name":     op.Name,
		}, server.ID, game)
	}

	// Notify that this server is online
	for _, sub := range s.statusSubs {
		sub(server.ID, "Online")
	}

	return nil
}

func (s *rconService) startPlayerListPolling(serverID int64, game domain.Game) {
	// Set up prevPlayers map for this server
	s.prevPlayers[serverID] = map[string]*onlinePlayer{}

	for {
		time.Sleep(game.GetConfig().PlayerListPollingInterval)

		client := s.clients[serverID]
		if client == nil {
			s.logger.Warn("Player list polling routine could not get the RCON client for this server",
				zap.Int64("Server ID", serverID))
			s.logger.Info("Exiting player list polling routine for server", zap.Int64("Server ID", serverID))
			return
		}

		players, err := s.getOnlinePlayers(serverID, game)
		if err != nil {
			s.logger.Warn("Player list polling routine not get online players for server", zap.Int64("Server ID", serverID))
			continue
		}

		onlinePlayers := map[string]*onlinePlayer{}
		for _, player := range players {
			onlinePlayers[player.PlayerID] = player
		}

		prevPlayers := s.prevPlayers[serverID]

		// Check for new player joins
		for playerGameID, player := range onlinePlayers {
			if prevPlayers[playerGameID] == nil {
				prevPlayers[playerGameID] = player

				// Player was not online previously so handle join
				s.HandlePlayerJoin(broadcast.Fields{
					"PlayerID": player.PlayerID,
					"Platform": game.GetPlatform().GetName(),
					"Name":     player.Name,
				}, serverID, game)
			}
		}

		// Check for existing player quits
		for prevPlayerGameID, prevPlayer := range prevPlayers {
			if onlinePlayers[prevPlayerGameID] == nil {
				delete(prevPlayers, prevPlayerGameID)

				// Player quit so broadcast quit
				for _, sub := range s.quitSubs {
					sub(broadcast.Fields{
						"PlayerID": prevPlayer.PlayerID,
						"Platform": game.GetPlatform().GetName(),
						"Name":     prevPlayer.Name,
					}, serverID, game)
				}
			}
		}

		// Update prevPlayers for this server
		s.prevPlayers[serverID] = prevPlayers
	}
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
		case broadcast.TypeJoin:
			s.HandlePlayerJoin(bcast.Fields, serverID, game)
			break
		case broadcast.TypeQuit:
			for _, sub := range s.quitSubs {
				sub(bcast.Fields, serverID, game)
			}
			break
		case broadcast.TypeChat:
			fields := bcast.Fields

			msgBody := &domain.ChatReceiveBody{
				ServerID:   serverID,
				PlayerID:   fields["PlayerID"],
				Platform:   game.GetPlatform().GetName(),
				Name:       fields["Name"],
				Message:    fields["Message"],
				SentByUser: false,
			}

			for _, sub := range s.chatSubs {
				sub(msgBody, serverID, game)
			}
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

func (s *rconService) SendChatMessage(body *domain.ChatSendBody) {
	client := s.clients[body.ServerID]

	// Check if this client's game has chat and RCON enabled
	conf := client.GetGame().GetConfig()

	if !conf.UseRCON || !conf.EnableChat {
		return
	}

	// If RCON and chat is enabled, then send the message
	command := fmt.Sprintf(client.GetGame().GetBroadcastCommand(), fmt.Sprintf("[%s]: %s", body.Sender, body.Message))

	if _, err := client.ExecCommand(command); err != nil {
		s.logger.Error("Could not send user chat message over RCON",
			zap.String("Message", body.Message),
			zap.Int64("Server ID", body.ServerID),
			zap.Error(err))
		return
	}

	s.logger.Info("Chat message forwarded to server",
		zap.String("Sender Name", body.Sender),
		zap.String("Message", body.Message),
		zap.Int64("Server ID", body.ServerID))
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

func (s *rconService) SubscribeChat(sub domain.ChatReceiveSubscriber) {
	s.chatSubs = append(s.chatSubs, sub)
}

func (s *rconService) HandlePlayerJoin(fields broadcast.Fields, serverID int64, game domain.Game) {
	// Broadcast join
	for _, sub := range s.joinSubs {
		sub(fields, serverID, game)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*2)
	defer cancel()

	playerID := fields["PlayerID"]
	platform := game.GetPlatform().GetName()
	name := fields["Name"]

	s.checkForBannedPlayer(ctx, platform, playerID, name, serverID, game)
}

func (s *rconService) checkForBannedPlayer(ctx context.Context, platform, playerID, name string, serverID int64, game domain.Game) {
	// Check if this player should be banned
	isBanned, timeRemaining, err := s.infractionService.PlayerIsBanned(ctx, platform, playerID)
	if err != nil {
		s.logger.Error("Could not check if player is banned",
			zap.String("Player ID", playerID),
			zap.String("Platform", platform),
			zap.Error(err))
		return
	}

	// If this player is not supposed to be banned then return since this function is only here to check bans.
	if !isBanned {
		return
	}

	// Get ban command from game settings
	gameSettings, err := s.gameService.GetGameSettings(game)
	if err != nil {
		s.logger.Error("Could not get game settings from game repo",
			zap.String("Game name", game.GetName()),
			zap.Error(err))
		return
	}

	banCmd := gameSettings.BanCommandPattern

	// Replace placeholders inside banCmd with player data
	banCmd = strings.ReplaceAll(banCmd, "{{PLAYER_ID}}", playerID)
	banCmd = strings.ReplaceAll(banCmd, "{{PLAYER_NAME}}", name)
	banCmd = strings.ReplaceAll(banCmd, "{{DURATION}}", strconv.FormatInt(timeRemaining, 10))
	banCmd = strings.ReplaceAll(banCmd, "{{REASON}}", "Refractor Ban Synchronization")

	// Ban the player for the correct remainder
	client := s.GetServerClient(serverID)
	res, err := client.ExecCommand(banCmd)
	if err != nil {
		s.logger.Error("Could not execute ban command",
			zap.String("Player ID", playerID),
			zap.String("Platform", platform),
			zap.Int64("Server ID", serverID),
			zap.Error(err))
		return
	}

	s.logger.Info("Banned player from server",
		zap.Int64("Server ID", serverID),
		zap.String("Player ID", playerID),
		zap.String("Platform", platform),
		zap.String("Reason For Ban", "Non-expired ban infraction on player record"),
		zap.String("Response From Server", res))
}
