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
	"Refractor/pkg/perms"
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
	authorizer      domain.Authorizer
	timeout         time.Duration
	logger          *zap.Logger
	infractionTypes map[string]domain.InfractionType
}

func NewInfractionService(repo domain.InfractionRepo, pr domain.PlayerRepo, sr domain.ServerRepo, a domain.Authorizer,
	to time.Duration, log *zap.Logger) domain.InfractionService {
	return &infractionService{
		repo:            repo,
		playerRepo:      pr,
		serverRepo:      sr,
		authorizer:      a,
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

// Update updates an infraction. If a user is set inside the passed in context with the key "user" then that user's
// permission to update the target infraction is checked. Otherwise, calls to this function are seen as trusted and
// are not authorized.
//
// When allowing this function to be executed by user requests, make sure they are authorized by setting the user in
// context under the key "user".
func (s *infractionService) Update(c context.Context, id int64, args domain.UpdateArgs) (*domain.Infraction, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// Get infraction which will be modified
	infraction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if the user is present in the passed in context. If they are, run permission checks. Otherwise, assume this
	// service call was not caused by a user and does not need to be authorized.
	user, ok := ctx.Value("user").(*domain.AuthUser)
	if ok {
		hasPermission, err := s.hasUpdatePermissions(ctx, infraction, user)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, domain.NewHTTPError(nil, http.StatusUnauthorized,
				"You do not have permission to update this infraction.")
		}
	}

	// Get filtered args
	args, err = s.filterUpdateArgs(infraction, args)
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

func (s *infractionService) hasUpdatePermissions(ctx context.Context, infraction *domain.Infraction, user *domain.AuthUser) (bool, error) {
	// The user will be granted permission to update this infraction if any of the following paths are satisfied:
	// 1. The user is an admin or super admin
	// OR:
	// 1. The target infraction was created by the user
	// 2. The user has permission to edit infraction records created by them.
	// OR:
	// 1. The user has permission to edit any infraction, even those not created by them.

	// Get computed user permissions for the given server
	userPerms, err := s.authorizer.GetPermissions(ctx, domain.AuthScope{
		Type: domain.AuthObjServer,
		ID:   infraction.ServerID,
	}, user.Identity.Id)
	if err != nil {
		return false, err
	}

	// Grant access if the user is an admin or super admin
	if userPerms.CheckFlag(perms.GetFlag(perms.FlagAdministrator)) || userPerms.CheckFlag(perms.GetFlag(perms.FlagSuperAdmin)) {
		return true, nil
	}

	// Grant access if the target infraction was created by the user and they have permission to edit their own infractions
	if infraction.UserID.Valid && infraction.UserID.ValueOrZero() == user.Identity.Id &&
		userPerms.CheckFlag(perms.GetFlag(perms.FlagEditOwnInfractions)) {
		return true, nil
	}

	// Grant access if the user has permission to edit any infraction
	if userPerms.CheckFlag(perms.GetFlag(perms.FlagEditAnyInfractions)) {
		return true, nil
	}

	// Otherwise, deny access
	return false, nil
}

// filterUpdateArgs filters the arguments to only include the allowed update fields of the target infraction type.
func (s *infractionService) filterUpdateArgs(infraction *domain.Infraction, args domain.UpdateArgs) (domain.UpdateArgs, error) {
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

func (s *infractionService) Delete(c context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// TODO: Check permissions

	return s.repo.Delete(ctx, id)
}
