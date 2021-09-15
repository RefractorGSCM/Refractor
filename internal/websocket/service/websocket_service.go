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
	pool       *websocket.Pool
	playerRepo domain.PlayerRepo
	authorizer domain.Authorizer
	timeout    time.Duration
	logger     *zap.Logger
}

func NewWebsocketService(pr domain.PlayerRepo, a domain.Authorizer, to time.Duration, log *zap.Logger) domain.WebsocketService {
	return &websocketService{
		pool:       websocket.NewPool(log),
		playerRepo: pr,
		authorizer: a,
		timeout:    to,
		logger:     log,
	}
}

func (s *websocketService) CreateClient(userID string, conn net.Conn) {
	client := websocket.NewClient(userID, conn, s.pool, s.logger)

	s.pool.Register <- client
	client.Read()
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

type playerJoinQuitData struct {
	ServerID int64  `json:"serverId"`
	PlayerID string `json:"id"`
	Platform string `json:"platform"`
	Name     string `json:"name"`
	Watched  bool   `json:"watched"`
}

func (s *websocketService) HandlePlayerJoin(fields broadcast.Fields, serverID int64, game domain.Game) {
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
		Type: "player-join",
		Body: playerJoinQuitData{
			ServerID: serverID,
			PlayerID: playerID,
			Platform: platform,
			Name:     foundPlayer.CurrentName,
			Watched:  foundPlayer.Watched,
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
			PlayerID: playerID,
			Platform: platform,
			Name:     foundPlayer.CurrentName,
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
