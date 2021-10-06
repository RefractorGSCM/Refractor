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
	"Refractor/pkg/broadcast"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type playerService struct {
	repo     domain.PlayerRepo
	nameRepo domain.PlayerNameRepo
	timeout  time.Duration
	logger   *zap.Logger
}

func NewPlayerService(repo domain.PlayerRepo, nameRepo domain.PlayerNameRepo, to time.Duration, log *zap.Logger) domain.PlayerService {
	return &playerService{
		repo:     repo,
		nameRepo: nameRepo,
		timeout:  to,
		logger:   log,
	}
}

func (s *playerService) HandlePlayerJoin(fields broadcast.Fields, serverID int64, game domain.Game) {
	ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
	defer cancel()

	playerID := fields["PlayerID"]
	platform := game.GetPlatform().GetName()
	name := fields["Name"]

	// Check if this player already exists
	foundPlayer, err := s.repo.GetByID(ctx, platform, playerID)
	if err != nil && errors.Cause(err) != domain.ErrNotFound {
		s.logger.Error("Could not get player by id",
			zap.String("PlayerID", playerID),
			zap.String("Platform", platform),
			zap.Error(err))
		return
	}

	// If foundPlayer is nil but the program didn't return from the above error check, then we know that
	// they don't exist so we create them.
	if foundPlayer == nil {
		newPlayer := &domain.Player{
			PlayerID:    playerID,
			Platform:    platform,
			CurrentName: name,
		}

		if err := s.repo.Store(ctx, newPlayer); err != nil {
			s.logger.Error("Could not store non existent player",
				zap.String("PlayerID", playerID),
				zap.String("Platform", platform),
				zap.Error(err))
			return
		}

		s.logger.Info("New player recorded",
			zap.String("PlayerID", playerID),
			zap.String("Platform", platform))
		return
	}

	// Otherwise, if the player already exists then check if their name has changed.
	if foundPlayer.CurrentName != name {
		s.logger.Info("Player name change detected",
			zap.String("PlayerID", playerID),
			zap.String("Platform", platform),
			zap.String("CurrentName", foundPlayer.CurrentName),
			zap.String("NewName", name))

		if err := s.nameRepo.UpdateName(ctx, foundPlayer, name); err != nil {
			s.logger.Error("Could not update player name",
				zap.String("PlayerID", playerID),
				zap.String("Platform", platform),
				zap.String("CurrentName", foundPlayer.CurrentName),
				zap.String("NewName", name),
				zap.Error(err))
			return
		}
	}

	// Update their LastSeen field to the current time
	if _, err := s.repo.Update(ctx, foundPlayer.Platform, foundPlayer.PlayerID, domain.UpdateArgs{
		"LastSeen": time.Now(),
	}); err != nil {
		s.logger.Error("Could not update player last seen field",
			zap.String("PlayerID", playerID),
			zap.String("Platform", platform),
			zap.Error(err))
		return
	}
}

func (s *playerService) HandlePlayerQuit(fields broadcast.Fields, serverID int64, game domain.Game) {
	ctx, cancel := context.WithTimeout(context.TODO(), s.timeout)
	defer cancel()

	playerID := fields["PlayerID"]
	platform := game.GetPlatform().GetName()

	// Check if this player already exists
	foundPlayer, err := s.repo.GetByID(ctx, platform, playerID)
	if err != nil && errors.Cause(err) != domain.ErrNotFound {
		s.logger.Error("Could not get player by id",
			zap.String("PlayerID", playerID),
			zap.String("Platform", platform),
			zap.Error(err))
		return
	}

	if foundPlayer == nil {
		s.logger.Warn("Found player was nil",
			zap.String("PlayerID", playerID),
			zap.String("Platform", platform),
			zap.Error(err))
		return
	}

	// Update their LastSeen field to the current time
	if _, err := s.repo.Update(ctx, foundPlayer.Platform, foundPlayer.PlayerID, domain.UpdateArgs{
		"LastSeen": time.Now(),
	}); err != nil {
		s.logger.Error("Could not update player last seen field",
			zap.String("PlayerID", playerID),
			zap.String("Platform", platform),
			zap.Error(err))
		return
	}
}

func (s *playerService) GetPlayer(c context.Context, id, platform string) (*domain.Player, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.GetByID(ctx, platform, id)
}
