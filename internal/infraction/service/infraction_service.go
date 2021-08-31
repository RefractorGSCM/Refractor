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
	"go.uber.org/zap"
	"time"
)

type infractionService struct {
	repo    domain.InfractionRepo
	timeout time.Duration
	logger  *zap.Logger
}

func NewInfractionService(repo domain.InfractionRepo, to time.Duration, log *zap.Logger) domain.InfractionService {
	return &infractionService{
		repo:    repo,
		timeout: to,
		logger:  log,
	}
}

func (s *infractionService) Store(c context.Context, infraction *domain.Infraction) (*domain.Infraction, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	infraction, err := s.repo.Store(ctx, infraction)
	if err != nil {
		return nil, err
	}

	return infraction, nil
}

func (s *infractionService) GetByID(c context.Context, id int64) (*domain.Infraction, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.GetByID(ctx, id)
}
