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
	"Refractor/authcheckers"
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
	attachmentRepo  domain.AttachmentRepo
	userMetaRepo    domain.UserMetaRepo
	authorizer      domain.Authorizer
	timeout         time.Duration
	logger          *zap.Logger
	infractionTypes map[string]domain.InfractionType
}

func NewInfractionService(repo domain.InfractionRepo, pr domain.PlayerRepo, sr domain.ServerRepo, ar domain.AttachmentRepo,
	umr domain.UserMetaRepo, a domain.Authorizer, to time.Duration, log *zap.Logger) domain.InfractionService {
	return &infractionService{
		repo:            repo,
		playerRepo:      pr,
		serverRepo:      sr,
		attachmentRepo:  ar,
		userMetaRepo:    umr,
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

func (s *infractionService) Store(c context.Context, infraction *domain.Infraction, attachments []*domain.Attachment) (*domain.Infraction, error) {
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

	// Create attachments
	for _, attachment := range attachments {
		attachment.InfractionID = infraction.InfractionID

		if err := s.attachmentRepo.Store(ctx, attachment); err != nil {
			s.logger.Error("Could not create attachment",
				zap.Int64("InfractionID", infraction.InfractionID),
				zap.String("Attachment URL", attachment.URL),
				zap.String("Attachment Note", attachment.Note),
				zap.Error(err),
			)

			// Do not fully return out of the function since attachment creation is not considered mission critical
			continue
		}
	}

	return infraction, nil
}

// GetByID returns an infraction with a matching ID.
//
// If a user is set inside the provided context with the key "user" then permissions are checked against the server
// this infraction was recorded on. If no user is provided in context, then authorization is skipped.
//
// GetByID also fetches the username of each issuing staff member for each infraction.
func (s *infractionService) GetByID(c context.Context, id int64) (*domain.Infraction, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	infraction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if a user exists in the context. If they do, check permissions.
	user, checkAuth := ctx.Value("user").(*domain.AuthUser)
	isAuthorized := false

	if checkAuth {
		// Check if the user has permission to view player records on the infraction's server.
		hasPermission, err := s.authorizer.HasPermission(ctx, domain.AuthScope{
			Type: domain.AuthObjServer,
			ID:   infraction.ServerID,
		}, user.Identity.Id, authcheckers.HasOneOfPermissions(true, perms.FlagViewPlayerRecords, perms.FlagViewInfractionRecords))
		if err != nil {
			return nil, err
		}

		// NOTE: It may seem a little backwards that we are checking both the FlagViewPlayerRecords or FlagViewInfractionRecords
		// to check whether the user is authorized to view an infraction record. Wouldn't solely checking FlagViewInfractionRecords
		// make more sense? On the surface, yes it would, but from an implementation standpoint this would complicate things.
		//
		// Take the following situation where only FlagViewInfractionRecords is checked:
		// 1. The player has permission to view player records
		// 2. They open a player's summary and see a list of partial infraction previews
		// 3. They click on one to view it's full info but oops! They are denied access because they only had
		//    FlagViewPlayerRecords.
		//
		// Since Infractions are tied so strongly to specific players, it makes more sense to allow access to users with
		// either of these permissions, even if it seems like a bit of a domain boundary violation. This is simpler.

		isAuthorized = hasPermission
	}

	if checkAuth && !isAuthorized {
		return nil, domain.NewHTTPError(nil, http.StatusUnauthorized,
			"You do not have permission to view this infraction.")
	}

	// Get issuer username
	username, err := s.userMetaRepo.GetUsername(ctx, infraction.UserID.ValueOrZero())
	if err != nil {
		s.logger.Error("Could not get infraction issuer's username",
			zap.Int64("Infraction ID", infraction.InfractionID),
			zap.String("User ID", infraction.UserID.ValueOrZero()),
			zap.Error(err),
		)
	}

	infraction.IssuerName = username

	return infraction, nil
}

// GetByPlayer returns all infractions for a player on a given platform.
//
// If a user is set inside the provided context with the key "user" then permissions are checked against each server
// so that only infractions belonging to servers the requesting user has access to are returned.
//
// GetByPlayer also fetches the username of each issuing staff member for each infraction.
func (s *infractionService) GetByPlayer(c context.Context, playerID, platform string) ([]*domain.Infraction, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	authorizedServers := map[int64]bool{}

	// Check if a user exists in the context. If they do, check permissions.
	user, checkAuth := ctx.Value("user").(*domain.AuthUser)

	// Get player infractions
	infractions, err := s.repo.GetByPlayer(ctx, playerID, platform)
	if err != nil {
		if errors.Cause(err) != domain.ErrNotFound {
			return nil, err
		}

		infractions = []*domain.Infraction{}
		checkAuth = false // no need to check auth if no infractions were found
	}

	if checkAuth {
		// Get all servers
		servers, err := s.serverRepo.GetAll(ctx)
		if err != nil {
			return nil, err
		}

		// Check if the user has permission to view player records on each server. If they do, then add the ID of this
		// server to the list of authorizedServers.
		for _, server := range servers {
			hasPermission, err := s.authorizer.HasPermission(ctx, domain.AuthScope{
				Type: domain.AuthObjServer,
				ID:   server.ID,
			}, user.Identity.Id, authcheckers.HasPermission(perms.FlagViewPlayerRecords, true))
			if err != nil {
				s.logger.Error("Could not check user auth for server",
					zap.String("User ID", user.Identity.Id),
					zap.Int64("Server ID", server.ID),
					zap.Error(err),
				)
				return nil, err
			}

			if hasPermission {
				authorizedServers[server.ID] = true
			} else {
				authorizedServers[server.ID] = false
			}
		}
	}

	var outputInfractions []*domain.Infraction
	if checkAuth {
		// Filter out infractions which belong to servers not in the authorizedServers slice.
		for _, i := range infractions {
			if authorizedServers[i.ServerID] {
				outputInfractions = append(outputInfractions, i)
			}
		}
	} else {
		outputInfractions = infractions
	}

	// Get username of issuer for each infraction and add to infraction object
	for _, infr := range outputInfractions {
		if !infr.UserID.Valid {
			continue
		}

		username, err := s.userMetaRepo.GetUsername(ctx, infr.UserID.ValueOrZero())
		if err != nil {
			s.logger.Error("Could not get infraction issuer's username",
				zap.Int64("Infraction ID", infr.InfractionID),
				zap.String("User ID", infr.UserID.ValueOrZero()),
				zap.Error(err),
			)
			continue
		}

		infr.IssuerName = username
	}

	return outputInfractions, nil
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

// Delete deletes an infraction. If a user is set inside the passed in context with the key "user" then that user's
// permission to delete the target infraction is checked. Otherwise, calls to this function are seen as trusted and
// are not authorized.
//
// When allowing this function to be executed by user requests, make sure they are authorized by setting the user in
// context under the key "user".
func (s *infractionService) Delete(c context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// Get infraction which will be modified
	infraction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if the user is present in the passed in context. If they are, run permission checks. Otherwise, assume this
	// service call was not caused by a user and does not need to be authorized.
	user, ok := ctx.Value("user").(*domain.AuthUser)
	if ok {
		hasPermission, err := s.hasDeletePermissions(ctx, infraction, user)
		if err != nil {
			return err
		}

		if !hasPermission {
			return domain.NewHTTPError(nil, http.StatusUnauthorized,
				"You do not have permission to delete this infraction.")
		}
	}

	return s.repo.Delete(ctx, id)
}

func (s *infractionService) hasDeletePermissions(ctx context.Context, infraction *domain.Infraction, user *domain.AuthUser) (bool, error) {
	// The user will be granted permission to delete this infraction if any of the following paths are satisfied:
	// 1. The user is an admin or super admin
	// OR:
	// 1. The target infraction was created by the user
	// 2. The user has permission to delete infraction records created by them.
	// OR:
	// 1. The user has permission to delete any infraction, even those not created by them.

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
		userPerms.CheckFlag(perms.GetFlag(perms.FlagDeleteOwnInfractions)) {
		return true, nil
	}

	// Grant access if the user has permission to edit any infraction
	if userPerms.CheckFlag(perms.GetFlag(perms.FlagDeleteAnyInfractions)) {
		return true, nil
	}

	// Otherwise, deny access
	return false, nil
}

func (s *infractionService) GetLinkedChatMessages(c context.Context, id int64) ([]*domain.ChatMessage, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.GetLinkedChatMessages(ctx, id)
}

func (s *infractionService) LinkChatMessage(c context.Context, id int64, messageID int64) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.LinkChatMessage(ctx, id, messageID)
}

func (s *infractionService) UnlinkChatMessage(c context.Context, id int64, messageID int64) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.UnlinkChatMessage(ctx, id, messageID)
}
