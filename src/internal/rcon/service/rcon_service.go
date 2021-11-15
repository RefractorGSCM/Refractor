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
	"sync"
	"time"
)

type rconService struct {
	logger        *zap.Logger
	clients       map[int64]domain.RCONClient
	gameService   domain.GameService
	serverRepo    domain.ServerRepo
	clientCreator domain.ClientCreator

	joinSubs       []domain.BroadcastSubscriber
	quitSubs       []domain.BroadcastSubscriber
	modActionSubs  []domain.BroadcastSubscriber
	playerListSubs []domain.PlayerListUpdateSubscriber
	statusSubs     []domain.ServerStatusSubscriber
	chatSubs       []domain.ChatReceiveSubscriber
	prevPlayers    map[int64]map[string]*domain.OnlinePlayer

	clientsLock sync.Mutex
	prevPlayersLock sync.Mutex
}

func NewRCONService(log *zap.Logger, gs domain.GameService, sr domain.ServerRepo) domain.RCONService {
	return &rconService{
		logger:         log,
		clients:        map[int64]domain.RCONClient{},
		serverRepo:     sr,
		clientCreator:  clientcreator.NewClientCreator(),
		gameService:    gs,
		joinSubs:       []domain.BroadcastSubscriber{},
		quitSubs:       []domain.BroadcastSubscriber{},
		playerListSubs: []domain.PlayerListUpdateSubscriber{},
		statusSubs:     []domain.ServerStatusSubscriber{},
		chatSubs:       []domain.ChatReceiveSubscriber{},
		prevPlayers:    map[int64]map[string]*domain.OnlinePlayer{},
	}
}

func (s *rconService) CreateClient(server *domain.Server) error {
	if !s.gameService.GameExists(server.Game) {
		return fmt.Errorf("could not create RCON client for servers with a non-existent game: %s", server.Game)
	}

	// Check if a client already exists. If one does, close the associated connections and delete it.
	s.clientsLock.Lock()
	currentClient := s.clients[server.ID]
	if currentClient != nil {
		_ = currentClient.Close()
		currentClient.WaitGroup().Wait() // wait for the client to disconnect
		delete(s.clients, server.ID)
	}
	s.clientsLock.Unlock()

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

	// Connect the client
	if err := client.Connect(); err != nil {
		return err
	}

	// Run init commands
	for _, cmd := range game.GetConfig().RCONInitCommands {
		if _, err := client.ExecCommand(cmd); err != nil {
			s.logger.Error("Could not execute RCON init command",
				zap.Int64("Server ID", server.ID),
				zap.String("Command", cmd),
				zap.Error(err))
			continue
		}
	}

	if gameConfig.PlayerListPollingEnabled() {
		go s.startPlayerListPolling(server.ID, game)
	}

	if gameConfig.PlayerListRefreshEnabled() {
		go s.startPlayerListRefreshPolling(server.ID, game)
	}

	// Add to list of clients
	s.clientsLock.Lock()
	s.clients[server.ID] = client
	s.clientsLock.Unlock()

	// Get currently online players
	onlinePlayers, err := s.getOnlinePlayers(server.ID, game)
	if err != nil {
		return err
	}

	// Dispatch player join events for all currently online players
	for _, op := range onlinePlayers {
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
	s.prevPlayersLock.Lock()
	s.prevPlayers[serverID] = map[string]*domain.OnlinePlayer{}
	s.prevPlayersLock.Unlock()

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

		onlinePlayers := map[string]*domain.OnlinePlayer{}
		for _, player := range players {
			onlinePlayers[player.PlayerID] = player
		}

		s.prevPlayersLock.Lock()
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
		s.prevPlayersLock.Unlock()
	}
}

// startPlayerListRefreshPolling is different from startPlayerListPolling. startPlayerListPolling is used as a primary
// method of keeping a player list up to date for servers which don't dispatch event broadcasts. This method is used to
// re-fetch the entire player list, ignoring any differences and just setting it.
//
// This is useful in case of server desyncs which can happen from time to time. e.g, if a server is killed without
// cleanly terminating the RCON connection or dispatching player quit notifications the player list will not be updated.
// This method helps mitigate that occurrence by periodically refreshing the entire player list.
func (s *rconService) startPlayerListRefreshPolling(serverID int64, game domain.Game) {
	for {
		time.Sleep(game.GetConfig().PlayerListRefreshInterval)

		if err := s.RefreshPlayerList(serverID, game); err != nil {
			s.logger.Error("Could not refresh player list from polling routine", zap.Error(err))
			continue
		}
	}
}

func (s *rconService) RefreshPlayerList(serverID int64, game domain.Game) error {
	// Ensure that server client exists
	if s.clients[serverID] == nil {
		return domain.ErrNotFound
	}

	// Get currently online players
	onlinePlayers, err := s.getOnlinePlayers(serverID, game)
	if err != nil {
		return err
	}

	// Broadcast full player list refresh to subscribers
	for _, sub := range s.playerListSubs {
		sub(serverID, onlinePlayers, game)
	}

	return nil
}

func (s *rconService) GetClients() map[int64]domain.RCONClient {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()
	return s.clients
}

func (s *rconService) GetServerClient(serverID int64) domain.RCONClient {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()
	return s.clients[serverID]
}

func (s *rconService) DeleteClient(serverID int64) {
	s.clientsLock.Lock()
	delete(s.clients, serverID)
	s.clientsLock.Unlock()
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
			break
			//case broadcast.TypeBan:
			//	for _, sub := range s.modActionSubs {
			//		sub(bcast.Fields, serverID, game)
			//	}
			//	break
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

func (s *rconService) getOnlinePlayers(serverID int64, game domain.Game) ([]*domain.OnlinePlayer, error) {
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

	var onlinePlayers []*domain.OnlinePlayer

	for _, player := range players {
		fields := regexutils.MapNamedMatches(playerListPattern, player)

		onlinePlayers = append(onlinePlayers, &domain.OnlinePlayer{
			PlayerID: fields["PlayerID"],
			Name:     fields["Name"],
		})
	}

	return onlinePlayers, nil
}

func (s *rconService) StartReconnectRoutine(serverID int64, data *domain.ServerData) {
	var delay = time.Second * 5

	var server *domain.Server
	for {
		time.Sleep(delay)

		// Check if client is already connected
		if s.clients[serverID] != nil {
			s.logger.Info("Server RCON client already connected", zap.Int64("Server ID", serverID))
			break
		}

		// Get updated server. We do this because it's possible that the server has been updated since this reconnect
		// service was started. Fetching the server each run isn't very expensive, and it lets us be sure that we're
		// connecting to the right place with the right settings!
		var err error
		server, err = s.serverRepo.GetByID(context.TODO(), serverID)
		if err != nil {
			s.logger.Warn(
				"RCON reconnect routine could not get server by ID",
				zap.Int64("Server", server.ID),
				zap.Error(err),
			)
			continue
		}

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

func (s *rconService) SubscribeModeratorAction(sub domain.BroadcastSubscriber) {
	s.modActionSubs = append(s.modActionSubs, sub)
}

func (s *rconService) HandlePlayerJoin(fields broadcast.Fields, serverID int64, game domain.Game) {
	// Broadcast join
	for _, sub := range s.joinSubs {
		sub(fields, serverID, game)
	}
}

func (s *rconService) HandleServerUpdate(server *domain.Server) {
	s.logger.Info("Received server update", zap.Int64("Server ID", server.ID))

	client := s.clients[server.ID]
	if client != nil {
		if err := client.Close(); err != nil {
			return
		}
		s.DeleteClient(server.ID)
	}

	// Notify disconnect
	for _, sub := range s.statusSubs {
		sub(server.ID, "Offline")
	}

	if err := s.CreateClient(server); err != nil {
		return
	}
}
