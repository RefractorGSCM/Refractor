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
	"Refractor/internal/infraction/types"
	"Refractor/pkg/whitelist"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type infractionService struct {
	repo            domain.InfractionRepo
	playerRepo      domain.PlayerRepo
	serverRepo      domain.ServerRepo
	timeout         time.Duration
	logger          *zap.Logger
	infractionTypes map[string]domain.InfractionType
}

func NewInfractionService(repo domain.InfractionRepo, pr domain.PlayerRepo, sr domain.ServerRepo, to time.Duration, log *zap.Logger) domain.InfractionService {
	return &infractionService{
		repo:            repo,
		playerRepo:      pr,
		serverRepo:      sr,
		timeout:         to,
		logger:          log,
		infractionTypes: getInfractionTypes(),
	}
}

func getInfractionTypes() map[string]domain.InfractionType {
	return map[string]domain.InfractionType{
		domain.InfractionTypeWarning: &types.Warning{},
		domain.InfractionTypeMute:    &types.Mute{},
		domain.InfractionTypeKick:    &types.Kick{},
		domain.InfractionTypeBan:     &types.Ban{},
	}
}

func (s *infractionService) Store(c context.Context, infraction *domain.Infraction) (*domain.Infraction, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// Ensure that player exists
	playerExists, err := s.playerRepo.Exists(ctx, domain.FindArgs{
		"PlayerID": infraction.PlayerID,
		"Platform": infraction.Platform,
	})
	if err != nil {
		return nil, err
	}

	if !playerExists {
		return nil, &domain.HTTPError{
			Cause:   nil,
			Message: "Player not found",
			ValidationErrors: map[string]string{
				"player_id": "player not found",
			},
			Status: http.StatusNotFound,
		}
	}

	// Ensure the server exists
	serverExists, err := s.serverRepo.Exists(ctx, domain.FindArgs{
		"ServerID": infraction.ServerID,
	})
	if err != nil {
		return nil, err
	}

	if !serverExists {
		return nil, &domain.HTTPError{
			Cause:   nil,
			Message: "Server not found",
			ValidationErrors: map[string]string{
				"server_id": "server not found",
			},
			Status: http.StatusNotFound,
		}
	}

	infraction, err = s.repo.Store(ctx, infraction)
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

func (s *infractionService) Update(c context.Context, id int64, args domain.UpdateArgs) (*domain.Infraction, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// Get infraction which will be modified
	infraction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: permission checks

	// Get filtered args
	args, err = s.filterUpdateArgs(ctx, infraction, args)
	if err != nil {
		return nil, err
	}

	if len(args) < 1 {
		return nil, &domain.HTTPError{
			Success:          false,
			Message:          "No updatable fields were provided",
			ValidationErrors: nil,
			Status:           http.StatusBadRequest,
		}
	}

	// Update the infraction
	return s.repo.Update(ctx, id, args)
}

// filterUpdateArgs filters the arguments to only include the allowed update fields of the target infraction type.
func (s *infractionService) filterUpdateArgs(ctx context.Context, infraction *domain.Infraction, args domain.UpdateArgs) (domain.UpdateArgs, error) {
	// Get allowed update fields from the infraction type to determine whitelist
	infractionType := s.infractionTypes[infraction.Type]
	if infractionType == nil {
		s.logger.Warn("An attempt was made to update an infraction with an unknown type", zap.String("Type", infraction.Type))
		return nil, errors.New("invalid infraction type")
	}

	// Create a whitelist from the allowed update fields of this infraction type
	wl := whitelist.StringKeyMap(infractionType.AllowedUpdateFields())

	// Filter update args with whitelist
	args = wl.FilterKeys(args)

	return args, nil
}
