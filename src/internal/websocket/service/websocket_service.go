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
	"Refractor/authcheckers"
	"Refractor/domain"
	"Refractor/pkg/broadcast"
	"Refractor/pkg/websocket"
	"context"
	"go.uber.org/zap"
	"net"
	"time"
)

type websocketService struct {
	pool               *websocket.Pool
	playerRepo         domain.PlayerRepo
	userMetaRepo       domain.UserMetaRepo
	playerStatsService domain.PlayerStatsService
	gameService        domain.GameService
	authorizer         domain.Authorizer
	timeout            time.Duration
	logger             *zap.Logger
	chatSendSubs       []domain.ChatSendSubscriber
}

func NewWebsocketService(pr domain.PlayerRepo, umr domain.UserMetaRepo, pss domain.PlayerStatsService, gs domain.GameService,
	a domain.Authorizer, to time.Duration, log *zap.Logger) domain.WebsocketService {
	return &websocketService{
		pool:               websocket.NewPool(log),
		playerRepo:         pr,
		userMetaRepo:       umr,
		playerStatsService: pss,
		gameService:        gs,
		authorizer:         a,
		timeout:            to,
		logger:             log,
		chatSendSubs:       []domain.ChatSendSubscriber{},
	}
}

func (s *websocketService) CreateClient(userID string, conn net.Conn) {
	client := websocket.NewClient(userID, conn, s.pool, s.sendChatHandler, s.logger)

	s.pool.Register <- client
	client.Read()
}

func (s *websocketService) sendChatHandler(body *websocket.SendChatBody) {
	ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
	defer cancel()

	// get user's name
	username, err := s.userMetaRepo.GetUsername(ctx, body.UserID)
	if err != nil {
		s.logger.Error("Could not get user username", zap.String("User ID", body.UserID), zap.Error(err))
		return
	}

	transformed := &domain.ChatSendBody{
		ServerID:   body.ServerID,
		Message:    body.Message,
		Sender:     username,
		SentByUser: true,
	}

	for _, sub := range s.chatSendSubs {
		sub(transformed)
	}
}

func (s *websocketService) StartPool() {
	s.pool.Start()
}

func (s *websocketService) Broadcast(message *domain.WebsocketMessage) {
	s.pool.Broadcast <- message
}

func (s *websocketService) BroadcastServerMessage(message *domain.WebsocketMessage, serverID int64, authChecker domain.AuthChecker) error {
	for _, client := range s.pool.Clients {
		hasPermission, err := s.authorizer.HasPermission(context.TODO(), domain.AuthScope{
			Type: domain.AuthObjServer,
			ID:   serverID,
		}, client.UserID, authChecker)
		if err != nil {
			return err
		}

		if hasPermission {
			s.pool.SendDirect <- &domain.WebsocketDirectMessage{
				ClientID: client.ID,
				Message:  message,
			}
		}
	}

	return nil
}

func (s *websocketService) SendDirectMessage(message *domain.WebsocketMessage, userID string) {
	for _, client := range s.pool.Clients {
		if client.UserID == userID {
			s.pool.SendDirect <- &domain.WebsocketDirectMessage{
				ClientID: client.ID,
				Message:  message,
			}
		}
	}
}

type playerJoinQuitData struct {
	ServerID int64 `json:"serverId"`
	*domain.PlayerPayload
}

func (s *websocketService) HandlePlayerJoin(fields broadcast.Fields, serverID int64, game domain.Game) {
	ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
	defer cancel()

	platform := game.GetPlatform().GetName()
	playerID := fields["PlayerID"]

	playerPayload, err := s.playerStatsService.GetPlayerPayload(ctx, platform, playerID, game)
	if err != nil {
		s.logger.Error("Could not get player payload",
			zap.String("Platform", platform),
			zap.String("Player ID", playerID),
			zap.Error(err))
		return
	}

	if err := s.BroadcastServerMessage(&domain.WebsocketMessage{
		Type: "player-join",
		Body: playerJoinQuitData{
			ServerID:      serverID,
			PlayerPayload: playerPayload,
		},
	}, serverID, authcheckers.CanViewServer); err != nil {
		s.logger.Warn("Could not broadcast player join message",
			zap.Error(err))
		return
	}
}

func (s *websocketService) HandlePlayerQuit(fields broadcast.Fields, serverID int64, game domain.Game) {
	ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
	defer cancel()

	platform := game.GetPlatform().GetName()
	playerID := fields["PlayerID"]

	foundPlayer, err := s.playerRepo.GetByID(ctx, game.GetPlatform().GetName(), fields["PlayerID"])
	if err != nil {
		s.logger.Warn("Could not get player by ID",
			zap.String("PlayerID", playerID),
			zap.String("Platform", platform),
			zap.Error(err))
		return
	}

	if err := s.BroadcastServerMessage(&domain.WebsocketMessage{
		Type: "player-quit",
		Body: playerJoinQuitData{
			ServerID: serverID,
			PlayerPayload: &domain.PlayerPayload{
				Player: &domain.Player{
					PlayerID:    playerID,
					Platform:    platform,
					CurrentName: foundPlayer.CurrentName,
				},
			},
		},
	}, serverID, authcheckers.CanViewServer); err != nil {
		s.logger.Warn("Could not broadcast player quit message",
			zap.Error(err))
		return
	}
}

type serverStatusBody struct {
	ServerID int64  `json:"server_id"`
	Status   string `json:"status"`
}

func (s *websocketService) HandleServerStatusChange(serverID int64, status string) {
	if err := s.BroadcastServerMessage(&domain.WebsocketMessage{
		Type: "server-status",
		Body: serverStatusBody{
			ServerID: serverID,
			Status:   status,
		},
	}, serverID, authcheckers.CanViewServer); err != nil {
		s.logger.Warn("Could not broadcast server status message", zap.Error(err))
		return
	}
}

type playerListRefreshData struct {
	ServerID      int64                 `json:"server_id"`
	OnlinePlayers []*playerJoinQuitData `json:"online_players"`
}

func (s *websocketService) HandlePlayerListUpdate(serverID int64, players []*domain.OnlinePlayer, game domain.Game) {
	ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
	defer cancel()

	platform := game.GetPlatform().GetName()
	playerData := make([]*playerJoinQuitData, 0)

	for _, op := range players {
		playerPayload, err := s.playerStatsService.GetPlayerPayload(ctx, platform, op.PlayerID, game)
		if err != nil {
			s.logger.Error("Could not get player payload",
				zap.String("Platform", platform),
				zap.String("Player ID", op.PlayerID),
				zap.Error(err))
			return
		}

		playerData = append(playerData, &playerJoinQuitData{
			ServerID:      serverID,
			PlayerPayload: playerPayload,
		})
	}

	// Broadcast player data
	if err := s.BroadcastServerMessage(&domain.WebsocketMessage{
		Type: "player-list-refresh",
		Body: &playerListRefreshData{
			ServerID:      serverID,
			OnlinePlayers: playerData,
		},
	}, serverID, authcheckers.CanViewServer); err != nil {
		s.logger.Warn("Could not broadcast server player list refresh", zap.Error(err))
		return
	}
}

type infractionBody struct {
	InfractionID int64  `json:"id"`
	ServerID     int64  `json:"server_id"`
	Platform     string `json:"platform"`
	PlayerID     string `json:"player_id"`
	Type         string `json:"type"`
	Reason       string `json:"reason,omitempty"`
	Duration     int64  `json:"duration,omitempty"`
}

func (s *websocketService) HandleInfractionCreate(infraction *domain.Infraction) {
	s.Broadcast(&domain.WebsocketMessage{
		Type: "infraction-create",
		Body: &infractionBody{
			InfractionID: infraction.InfractionID,
			ServerID:     infraction.ServerID,
			Platform:     infraction.Platform,
			PlayerID:     infraction.PlayerID,
			Type:         infraction.Type,
			Reason:       infraction.Reason.ValueOrZero(),
			Duration:     infraction.Duration.ValueOrZero(),
		},
	})
}

func (s *websocketService) SubscribeChatSend(sub domain.ChatSendSubscriber) {
	s.chatSendSubs = append(s.chatSendSubs, sub)
}
