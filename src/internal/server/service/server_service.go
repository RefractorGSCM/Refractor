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
	"Refractor/pkg/bitperms"
	"Refractor/pkg/broadcast"
	"Refractor/pkg/perms"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type serverService struct {
	repo               domain.ServerRepo
	playerRepo         domain.PlayerRepo
	playerStatsService domain.PlayerStatsService
	authorizer         domain.Authorizer
	timeout            time.Duration
	logger             *zap.Logger
	serverData         map[int64]*domain.ServerData
}

func NewServerService(repo domain.ServerRepo, pr domain.PlayerRepo, pss domain.PlayerStatsService,
	a domain.Authorizer, timeout time.Duration, log *zap.Logger) domain.ServerService {
	return &serverService{
		repo:               repo,
		playerRepo:         pr,
		playerStatsService: pss,
		authorizer:         a,
		timeout:            timeout,
		logger:             log,
		serverData:         map[int64]*domain.ServerData{},
	}
}

func (s *serverService) Store(c context.Context, server *domain.Server) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	if err := s.repo.Store(ctx, server); err != nil {
		return err
	}

	if err := s.CreateServerData(server.ID); err != nil {
		return err
	}

	return nil
}

type serverResponse struct {
	Data domain.ServerData `json:"data"`
	*domain.Server
}

func (s *serverService) GetByID(c context.Context, id int64) (*domain.Server, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.GetByID(ctx, id)
}

func (s *serverService) GetAll(c context.Context) ([]*domain.Server, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	allServers, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return allServers, nil
}

func (s *serverService) GetAllAccessible(c context.Context) ([]*domain.Server, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	user, ok := c.Value("user").(*domain.AuthUser)
	if !ok || user == nil {
		return nil, fmt.Errorf("no user or invalid user found in context")
	}

	allServers, err := s.repo.GetAll(ctx)
	if err != nil {
		if errors.Cause(err) == domain.ErrNotFound {
			return []*domain.Server{}, nil
		}

		return nil, err
	}

	var results []*domain.Server

	// Filter out servers the user does not have access to view or servers which are deactivated
	for _, server := range allServers {
		// Filter out deactivated servers
		if server.Deactivated {
			continue
		}

		hasPermission, err := s.authorizer.HasPermission(ctx, domain.AuthScope{
			Type: domain.AuthObjServer,
			ID:   server.ID,
		}, user.Identity.Id, func(permissions *bitperms.Permissions) (bool, error) {
			hasPerm := permissions.CheckFlag(perms.GetFlag(perms.FlagViewServers))
			if hasPerm {
				return hasPerm, nil
			}

			hasPerm = permissions.CheckFlag(perms.GetFlag(perms.FlagAdministrator))
			if hasPerm {
				return hasPerm, nil
			}

			hasPerm = permissions.CheckFlag(perms.GetFlag(perms.FlagSuperAdmin))
			if hasPerm {
				return hasPerm, nil
			}

			return false, nil
		})

		if err != nil {
			s.logger.Error(
				"Could not check if user has permission to view server",
				zap.String("User ID", user.Identity.Id),
				zap.Int64("Server ID", server.ID),
				zap.Error(err),
			)
			return nil, err
		}

		// If the user has permission, add it to the results slice
		if hasPermission {
			results = append(results, server)
		} else {
			// If the user does not have permission to outright view the server, check if they have permission to view
			// player or infraction records on this server. If they do, append an incomplete server fragment to the list
			// containing only the ID and Name while keeping any more sensitive info hidden.
			hasPermission, err := s.authorizer.HasPermission(ctx, domain.AuthScope{
				Type: domain.AuthObjServer,
				ID:   server.ID,
			}, user.Identity.Id, authcheckers.HasOneOfPermissions(true, perms.FlagViewPlayerRecords, perms.FlagViewInfractionRecords))
			if err != nil {
				s.logger.Error(
					"Could not check if user has permission to view player/infraction records on server",
					zap.String("User ID", user.Identity.Id),
					zap.Int64("Server ID", server.ID),
					zap.Error(err),
				)
				return nil, err
			}

			if hasPermission {
				results = append(results, &domain.Server{
					ID:         server.ID,
					Name:       server.Name,
					IsFragment: true,
				})
			}
		}
	}

	return results, nil
}

func (s *serverService) Deactivate(c context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Deactivate(ctx, id)
}

func (s *serverService) Update(c context.Context, id int64, args domain.UpdateArgs) (*domain.Server, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Update(ctx, id, args)
}

func (s *serverService) CreateServerData(id int64) error {
	s.serverData[id] = &domain.ServerData{
		NeedsUpdate:   true,
		ServerID:      id,
		PlayerCount:   0,
		OnlinePlayers: map[string]*domain.Player{},
		Status:        "Unknown",
	}

	return nil
}

func (s *serverService) GetAllServerData() ([]*domain.ServerData, error) {
	var allData []*domain.ServerData

	for _, data := range s.serverData {
		// Get infraction counts for each player
		for _, p := range data.OnlinePlayers {
			count, err := s.playerStatsService.GetInfractionCount(context.TODO(), p.Platform, p.PlayerID)
			if err != nil {
				s.logger.Error("Could not get player infraction count for online player",
					zap.String("Platform", p.Platform),
					zap.String("Player ID", p.PlayerID),
					zap.Error(err))
				continue
			}

			p.InfractionCount = count
		}

		allData = append(allData, data)
	}

	return allData, nil
}

func (s *serverService) GetServerData(id int64) (*domain.ServerData, error) {
	data := s.serverData[id]

	if data == nil {
		return nil, fmt.Errorf("server data not found")
	}

	// Get infraction counts for each player
	for _, p := range data.OnlinePlayers {
		count, err := s.playerStatsService.GetInfractionCount(context.TODO(), p.Platform, p.PlayerID)
		if err != nil {
			s.logger.Error("Could not get player infraction count for online player",
				zap.String("Platform", p.Platform),
				zap.String("Player ID", p.PlayerID),
				zap.Error(err))
			continue
		}

		p.InfractionCount = count
	}

	return data, nil
}

func (s *serverService) HandlePlayerJoin(fields broadcast.Fields, serverID int64, game domain.Game) {
	ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
	defer cancel()

	playerID := fields["PlayerID"]
	platform := game.GetPlatform().GetName()

	player, err := s.playerRepo.GetByID(ctx, platform, playerID)
	if err != nil {
		s.logger.Warn("Could not get player by ID",
			zap.String("PlayerID", playerID),
			zap.String("Platform", platform),
			zap.Error(err))
		return
	}

	// Add player to server data
	s.serverData[serverID].OnlinePlayers[playerID] = player
}

func (s *serverService) HandlePlayerQuit(fields broadcast.Fields, serverID int64, game domain.Game) {
	playerID := fields["PlayerID"]
	// Remove player from server data
	delete(s.serverData[serverID].OnlinePlayers, playerID)
}

func (s *serverService) HandleServerStatusChange(serverID int64, status string) {
	data, err := s.GetServerData(serverID)
	if err != nil {
		s.logger.Warn("Could not get server data", zap.Int64("Server ID", serverID), zap.Error(err))
		return
	}

	data.Status = status
}
