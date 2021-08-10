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
	"go.uber.org/zap"
)

type rconService struct {
	logger  *zap.Logger
	clients map[int64]*domain.RCONClient
}

func NewRCONService(log *zap.Logger) domain.RCONService {
	return &rconService{
		logger:  log,
		clients: map[int64]*domain.RCONClient{},
	}
}

func (s *rconService) CreateClient(server *domain.Server) error {
	panic("implement me")
}

func (s *rconService) GetClients() map[int64]*domain.RCONClient {
	panic("implement me")
}

func (s *rconService) DeleteClient(serverID int64) {
	panic("implement me")
}
