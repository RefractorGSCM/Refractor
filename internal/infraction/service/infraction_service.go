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
	"net/http"
	"time"
)

type infractionService struct {
	repo       domain.InfractionRepo
	playerRepo domain.PlayerRepo
	serverRepo domain.ServerRepo
	timeout    time.Duration
	logger     *zap.Logger
}

func NewInfractionService(repo domain.InfractionRepo, pr domain.PlayerRepo, sr domain.ServerRepo, to time.Duration, log *zap.Logger) domain.InfractionService {
	return &infractionService{
		repo:       repo,
		playerRepo: pr,
		serverRepo: sr,
		timeout:    to,
		logger:     log,
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
