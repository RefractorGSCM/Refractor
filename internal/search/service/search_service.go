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
	"Refractor/pkg/whitelist"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type searchService struct {
	playerRepo     domain.PlayerRepo
	playerNameRepo domain.PlayerNameRepo
	infractionRepo domain.InfractionRepo
	timeout        time.Duration
	logger         *zap.Logger
}

func NewSearchService(pr domain.PlayerRepo, pnr domain.PlayerNameRepo, ir domain.InfractionRepo,
	to time.Duration, log *zap.Logger) domain.SearchService {
	return &searchService{
		playerRepo:     pr,
		playerNameRepo: pnr,
		infractionRepo: ir,
		timeout:        to,
		logger:         log,
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
			if errors.Cause(err) != domain.ErrNotFound {
				s.logger.Error("Could not get player by id",
					zap.String("Platform", platform),
					zap.String("Player ID", term),
					zap.Error(err),
				)

				return 0, nil, err
			}

			return 0, []*domain.Player{}, nil
		}

		return 1, []*domain.Player{result}, nil
	}

	return 0, nil, errors.New("unknown search type")
}

func (s searchService) SearchInfractions(c context.Context, args domain.FindArgs, limit, offset int) (int, []*domain.Infraction, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// Filter out illegal values
	wl := whitelist.StringKeyMap([]string{"Type", "Game", "PlayerID", "Platform", "ServerID", "UserID"})
	args = wl.FilterKeys(args)

	if len(args) == 0 {
		return 0, []*domain.Infraction{}, &domain.HTTPError{
			Success:          false,
			Message:          "No search fields were provided",
			ValidationErrors: nil,
			Status:           http.StatusBadRequest,
		}
	}

	// Execute search
	count, infractions, err := s.infractionRepo.Search(ctx, args, limit, offset)
	if err != nil {
		if err == domain.ErrNotFound {
			return 0, []*domain.Infraction{}, nil
		}

		s.logger.Error("Could not search infractions", zap.Error(err))
		return 0, []*domain.Infraction{}, err
	}

	// Get player name for each infraction
	for _, infraction := range infractions {
		currentName, _, err := s.playerNameRepo.GetNames(ctx, infraction.PlayerID, infraction.Platform)
		if err != nil {
			s.logger.Error(
				"Could not get player name for infraction",
				zap.Int64("Infraction ID", infraction.InfractionID),
				zap.String("Platform", infraction.Platform),
				zap.String("Player ID", infraction.PlayerID),
				zap.Error(err),
			)
		}

		infraction.PlayerName = currentName
	}

	return count, infractions, nil
}
