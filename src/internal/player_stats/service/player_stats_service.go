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
	playerRepo     domain.PlayerRepo
	infractionRepo domain.InfractionRepo
	gameService    domain.GameService
	timeout        time.Duration
	logger         *zap.Logger
}

func NewPlayerStatsService(pr domain.PlayerRepo, ir domain.InfractionRepo, gs domain.GameService, to time.Duration,
	log *zap.Logger) domain.PlayerStatsService {
	return &pStatService{
		playerRepo:     pr,
		infractionRepo: ir,
		gameService:    gs,
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

func (s *pStatService) GetPlayerPayload(c context.Context, platform, playerID string, game domain.Game) (*domain.PlayerPayload, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	settings, err := s.gameService.GetGameSettings(game)
	if err != nil {
		return nil, err
	}

	foundPlayer, err := s.playerRepo.GetByID(ctx, platform, playerID)
	if err != nil {
		s.logger.Error("Could not get player by ID",
			zap.String("PlayerID", playerID),
			zap.String("Platform", platform),
			zap.Error(err))
		return nil, err
	}

	// Get player infraction count
	infractionCount, err := s.infractionRepo.GetPlayerTotalInfractions(ctx, foundPlayer.Platform, foundPlayer.PlayerID)
	if err != nil {
		s.logger.Error("Could not get player infraction count",
			zap.String("Platform", foundPlayer.Platform),
			zap.String("Player ID", foundPlayer.PlayerID),
			zap.Error(err))
		return nil, err
	}

	// Get player infraction count in timespan
	infractionCountSinceTimespan, err := s.GetInfractionCountSince(ctx, platform, playerID,
		settings.General.PlayerInfractionTimespan)
	if err != nil {
		s.logger.Error("Could not get player infractions since configured timespan", zap.Error(err))
		return nil, err
	}

	return &domain.PlayerPayload{
		Player:                       foundPlayer,
		InfractionCount:              infractionCount,
		InfractionCountSinceTimespan: infractionCountSinceTimespan,
	}, nil
}
