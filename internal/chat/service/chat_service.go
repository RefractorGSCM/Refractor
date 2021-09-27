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
	"context"
	"go.uber.org/zap"
	"time"
)

type chatService struct {
	repo             domain.ChatRepo
	playerRepo       domain.PlayerRepo
	websocketService domain.WebsocketService
	timeout          time.Duration
	logger           *zap.Logger
}

func NewChatService(repo domain.ChatRepo, pr domain.PlayerRepo, wss domain.WebsocketService,
	to time.Duration, log *zap.Logger) domain.ChatService {
	return &chatService{
		repo:             repo,
		playerRepo:       pr,
		websocketService: wss,
		timeout:          to,
		logger:           log,
	}
}

func (s *chatService) Store(c context.Context, message *domain.ChatMessage) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Store(ctx, message)
}

func (s *chatService) HandleChatReceive(body *domain.ChatReceiveBody, serverID int64, game domain.Game) {
	ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
	defer cancel()

	// Broadcast message to websocket clients
	if err := s.websocketService.BroadcastServerMessage(&domain.WebsocketMessage{
		Type: "chat",
		Body: body,
	}, serverID, authcheckers.CanViewServer); err != nil {
		s.logger.Warn("Could not broadcast received chat message to websocket clients",
			zap.Int64("Server ID", body.ServerID),
			zap.String("Player ID", body.PlayerID),
			zap.String("Platform", body.Platform),
			zap.String("Name", body.Name),
			zap.String("Message", body.Message),
			zap.Bool("Sent By User", body.SentByUser),
			zap.Error(err),
		)

		// do not return as this is not a critical error
	}

	// Get the player who sent this message to make sure that they exist. If they don't or any other error occurs, we
	// skip logging this message.
	player, err := s.playerRepo.GetByID(ctx, body.Platform, body.PlayerID)
	if err != nil {
		s.logger.Error("Could not get player who sent a chat message",
			zap.String("Player ID", body.PlayerID),
			zap.String("Platform", body.Platform),
			zap.Error(err),
		)
		return
	}

	// Log chat message
	if err := s.Store(ctx, &domain.ChatMessage{
		PlayerID: player.PlayerID,
		Platform: body.Platform,
		ServerID: serverID,
		Message:  body.Message,
		Flagged:  false,
	}); err != nil {
		s.logger.Error("Could not store chat message in repo",
			zap.Int64("Server ID", body.ServerID),
			zap.String("Player ID", body.PlayerID),
			zap.String("Platform", body.Platform),
			zap.String("Name", body.Name),
			zap.String("Message", body.Message),
			zap.Bool("Sent By User", body.SentByUser),
			zap.Error(err),
		)
	}
}
