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
	"Refractor/pkg/bitperms"
	"Refractor/pkg/perms"
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"
)

type serverService struct {
	repo       domain.ServerRepo
	authorizer domain.Authorizer
	timeout    time.Duration
	logger     *zap.Logger
	serverData map[int64]*domain.ServerData
}

func NewServerService(repo domain.ServerRepo, a domain.Authorizer, timeout time.Duration, log *zap.Logger) domain.ServerService {
	return &serverService{
		repo:       repo,
		authorizer: a,
		timeout:    timeout,
		logger:     log,
		serverData: map[int64]*domain.ServerData{},
	}
}

func (s *serverService) Store(c context.Context, server *domain.Server) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Store(ctx, server)
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
		return nil, err
	}

	var results []*domain.Server

	// Filter out servers the user does not have access to view
	for _, server := range allServers {
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
	}

	return nil
}

func (s *serverService) GetAllServerData() ([]*domain.ServerData, error) {
	var allData []*domain.ServerData

	for _, data := range s.serverData {
		allData = append(allData, data)
	}

	return allData, nil
}

func (s *serverService) GetServerData(id int64) (*domain.ServerData, error) {
	data := s.serverData[id]

	if data == nil {
		return nil, fmt.Errorf("server data not found")
	}

	return data, nil
}
