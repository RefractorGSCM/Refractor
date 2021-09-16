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
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type searchService struct {
	playerRepo domain.PlayerRepo
	timeout    time.Duration
	logger     *zap.Logger
}

func NewSearchService(pr domain.PlayerRepo, to time.Duration, log *zap.Logger) domain.SearchService {
	return &searchService{
		playerRepo: pr,
		timeout:    to,
		logger:     log,
	}
}

func (s searchService) SearchPlayers(c context.Context, term, searchType, platform string, limit, offset int) (int, []*domain.Player, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	switch searchType {
	case "name":
		totalResults, results, err := s.playerRepo.SearchByName(ctx, term, limit, offset)
		if err != nil {
			s.logger.Error("Could not search player by name",
				zap.String("Name", term),
				zap.Int("Limit", limit),
				zap.Int("Offset", offset),
				zap.Error(err),
			)
			return 0, nil, err
		}

		return totalResults, results, nil
	case "id":
		result, err := s.playerRepo.GetByID(ctx, platform, term)
		if err != nil {
			s.logger.Error("Could not get player by id",
				zap.String("Platform", platform),
				zap.String("Player ID", term),
				zap.Error(err),
			)
			return 0, nil, err
		}

		return 1, []*domain.Player{result}, nil
	}

	return 0, nil, errors.New("unknown search type")
}
