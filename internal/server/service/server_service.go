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
	"context"
	"fmt"
	"time"
)

type serverService struct {
	repo       domain.ServerRepo
	timeout    time.Duration
	serverData map[int64]*domain.ServerData
}

func NewServerService(repo domain.ServerRepo, timeout time.Duration) domain.ServerService {
	return &serverService{
		repo:       repo,
		timeout:    timeout,
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

	// TODO: filter out servers the user does not have access to (permission checks)

	return allServers, nil
}

func (s *serverService) Deactivate(c context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Deactivate(ctx, id)
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
