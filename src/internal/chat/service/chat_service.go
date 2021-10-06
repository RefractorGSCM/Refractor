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
	"Refractor/pkg/perms"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type chatService struct {
	repo               domain.ChatRepo
	playerRepo         domain.PlayerRepo
	playerNameRepo     domain.PlayerNameRepo
	serverService      domain.ServerService
	websocketService   domain.WebsocketService
	flaggedWordService domain.FlaggedWordService
	authorizer         domain.Authorizer
	timeout            time.Duration
	logger             *zap.Logger
}

func NewChatService(repo domain.ChatRepo, pr domain.PlayerRepo, pnr domain.PlayerNameRepo, ss domain.ServerService,
	wss domain.WebsocketService, fws domain.FlaggedWordService, a domain.Authorizer, to time.Duration, log *zap.Logger) domain.ChatService {
	return &chatService{
		repo:               repo,
		playerRepo:         pr,
		playerNameRepo:     pnr,
		serverService:      ss,
		websocketService:   wss,
		flaggedWordService: fws,
		authorizer:         a,
		timeout:            to,
		logger:             log,
	}
}

func (s *chatService) Store(c context.Context, message *domain.ChatMessage) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// Check if this message contains any flagged words
	shouldBeFlagged, err := s.flaggedWordService.MessageContainsFlaggedWord(ctx, message.Message)
	if err != nil {
		s.logger.Error("Could not check if message contains flagged word", zap.Error(err))
		// do not return as this is not a critical error and storing the chat message is more important than flagging it
	}

	message.Flagged = shouldBeFlagged

	return s.repo.Store(ctx, message)
}

func (s *chatService) HandleUserSendChat(body *domain.ChatSendBody) {
	if !body.SentByUser {
		return
	}

	s.websocketService.Broadcast(&domain.WebsocketMessage{
		Type: "chat",
		Body: &domain.ChatReceiveBody{
			ServerID:   body.ServerID,
			Name:       body.Sender,
			Message:    body.Message,
			SentByUser: body.SentByUser,
		},
	})
}

type sentChatMessage struct {
	MessageID int64 `json:"id"`
	*domain.ChatReceiveBody
}

func (s *chatService) HandleChatReceive(body *domain.ChatReceiveBody, serverID int64, game domain.Game) {
	ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
	defer cancel()

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

	message := &domain.ChatMessage{
		PlayerID: player.PlayerID,
		Platform: body.Platform,
		ServerID: serverID,
		Message:  body.Message,
		Flagged:  false,
	}

	// Log chat message
	if err := s.Store(ctx, message); err != nil {
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

	// Broadcast message to websocket clients
	if err := s.websocketService.BroadcastServerMessage(&domain.WebsocketMessage{
		Type: "chat",
		Body: sentChatMessage{
			MessageID:       message.MessageID,
			ChatReceiveBody: body,
		},
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
}

func (s *chatService) GetRecentByServer(c context.Context, serverID int64, count int) ([]*domain.ChatMessage, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
	defer cancel()

	results, err := s.repo.GetRecentByServer(ctx, serverID, count)
	if err != nil {
		if errors.Cause(err) == domain.ErrNotFound {
			return []*domain.ChatMessage{}, nil
		}

		return nil, err
	}

	for _, msg := range results {
		// Get player name for message
		currentName, _, err := s.playerNameRepo.GetNames(ctx, msg.PlayerID, msg.Platform)
		if err != nil {
			s.logger.Error("Could not get current name for recent chat message",
				zap.String("Platform", msg.Platform),
				zap.String("Player ID", msg.PlayerID),
				zap.Int64("Message ID", msg.MessageID),
				zap.Error(err))
			continue
		}

		msg.Name = currentName
	}

	return results, nil
}

// GetFlaggedMessages returns n (count) amount of recent flagged messages.
//
// If a user is provided in context under the key "user", the user will be authorized against servers by their ability
// to view chat records.
//
// If no user is provided, we assume this is a system call and skip authorization.
func (s *chatService) GetFlaggedMessages(c context.Context, count int) ([]*domain.ChatMessage, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// Get servers the user has permission to view chat records of
	allServers, err := s.serverService.GetAll(ctx)
	if err != nil {
		if errors.Cause(err) == domain.ErrNotFound {
			return []*domain.ChatMessage{}, nil
		}

		return nil, err
	}
	user, checkAuth := ctx.Value("user").(*domain.AuthUser)

	var authorizedServers []int64

	for _, server := range allServers {
		if server.Deactivated {
			continue
		}

		if checkAuth {
			hasPermission, err := s.authorizer.HasPermission(ctx, domain.AuthScope{
				Type: domain.AuthObjServer,
				ID:   server.ID,
			}, user.Identity.Id, authcheckers.HasPermission(perms.FlagViewChatRecords, true))
			if err != nil {
				s.logger.Error("Could not check if user has permission to view chat records on this server",
					zap.Error(err))
				return nil, err
			}

			if hasPermission {
				authorizedServers = append(authorizedServers, server.ID)
			}
			continue
		}

		// if we're not checking auth, just add server to the list
		authorizedServers = append(authorizedServers, server.ID)
	}

	results, err := s.repo.GetFlaggedMessages(ctx, count, authorizedServers)
	if err != nil {
		if errors.Cause(err) == domain.ErrNotFound {
			return []*domain.ChatMessage{}, nil
		}

		return nil, err
	}

	return results, nil
}
