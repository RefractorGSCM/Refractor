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

type pStatService struct {
	infractionRepo domain.InfractionRepo
	timeout        time.Duration
	logger         *zap.Logger
}

func NewPlayerStatsService(ir domain.InfractionRepo, to time.Duration, log *zap.Logger) domain.PlayerStatsService {
	return &pStatService{
		infractionRepo: ir,
		timeout:        to,
		logger:         log,
	}
}

func (s *pStatService) GetInfractionCount(c context.Context, platform, playerID string) (int, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.infractionRepo.GetPlayerTotalInfractions(ctx, platform, playerID)
}

func (s *pStatService) GetInfractionCountSince(c context.Context, platform, playerID string, sinceMinutes int) (int, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	sinceDate := time.Now().Add(time.Duration(-sinceMinutes) * time.Minute)

	return s.infractionRepo.GetPlayerInfractionCountSince(ctx, platform, playerID, sinceDate)
}
