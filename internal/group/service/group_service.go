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
	"Refractor/pkg/bitperms"
	"Refractor/pkg/perms"
	"Refractor/pkg/pointer"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type groupService struct {
	repo       domain.GroupRepo
	authorizer domain.Authorizer
	timeout    time.Duration
	logger     *zap.Logger
}

func NewGroupService(r domain.GroupRepo, a domain.Authorizer, to time.Duration, log *zap.Logger) domain.GroupService {
	return &groupService{
		repo:       r,
		authorizer: a,
		timeout:    to,
		logger:     log,
	}
}

func (s *groupService) Store(c context.Context, group *domain.Group) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// Check if the user is attempting to set the super admin flag
	permString, err := s.unsetSuperAdminFlag(group.Permissions)
	if err != nil {
		return err
	}

	group.Permissions = permString

	return s.repo.Store(ctx, group)
}

func (s *groupService) GetAll(c context.Context) ([]*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	groups, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Add base group to the results
	baseGroup, err := s.repo.GetBaseGroup(ctx)
	if err != nil {
		return nil, err
	}
	groups = append(groups, baseGroup)

	return domain.GroupSlice(groups).SortByPosition(), nil
}

func (s *groupService) GetByID(c context.Context, id int64) (*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.GetByID(ctx, id)
}

func (s *groupService) Delete(c context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Delete(ctx, id)
}

func (s *groupService) Update(c context.Context, id int64, args domain.UpdateArgs) (*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// Check if the user is attempting to set the super admin flag
	if args["Permissions"] != nil {
		permsArg := pointer.DePointer(args["Permissions"])
		permString, _ := permsArg.(string)
		permString, err := s.unsetSuperAdminFlag(permString)
		if err != nil {
			return nil, err
		}

		args["Permissions"] = permString
	}

	return s.repo.Update(ctx, id, args)
}

func (s *groupService) Reorder(c context.Context, newPositions []*domain.GroupReorderInfo) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Reorder(ctx, newPositions)
}

func (s *groupService) UpdateBase(c context.Context, args domain.UpdateArgs) (*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	currentBase, err := s.repo.GetBaseGroup(ctx)
	if err != nil {
		return nil, err
	}

	// Only allow the updating of Permissions and Color
	if args["Permissions"] != nil {
		updatedPermissions := args["Permissions"].(*string)
		currentBase.Permissions = *updatedPermissions
	}

	if args["Color"] != nil {
		updatedColor := args["Color"].(*int)
		currentBase.Color = *updatedColor
	}

	// Set the base group
	if err := s.repo.SetBaseGroup(ctx, currentBase); err != nil {
		return nil, err
	}

	return currentBase, nil
}

func (s *groupService) AddUserGroup(c context.Context, groupctx domain.GroupSetContext) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	canSetGroup, err := s.canSetGroup(ctx, groupctx)
	if err != nil {
		return err
	}

	if !canSetGroup {
		return domain.NewHTTPError(nil, http.StatusUnauthorized,
			"You do not have permission to assign that group to that user.")
	}

	if err := s.repo.AddUserGroup(ctx, groupctx.TargetUserID, groupctx.GroupID); err != nil {
		if errors.Cause(err) == domain.ErrConflict {
			return domain.NewHTTPError(err, http.StatusConflict, "User already has this group")
		}
	}

	return nil
}

func (s *groupService) unsetSuperAdminFlag(permString string) (string, error) {
	if permString == "" {
		return permString, nil
	}

	newPerms, err := bitperms.FromString(permString)
	if err != nil {
		return "", err
	}

	superAdminFlag := perms.GetFlag(perms.FlagSuperAdmin)

	// If the super admin flag is set, disable it.
	if newPerms.CheckFlag(superAdminFlag) {
		newPerms = newPerms.UnsetFlag(superAdminFlag)

		return newPerms.String(), nil
	}

	// Otherwise, simply return the old value since the super admin flag is not set.
	return permString, nil
}

func (s *groupService) canSetGroup(ctx context.Context, groupctx domain.GroupSetContext) (bool, error) {
	// A user can only add/remove a group to another user if the following criteria is met:
	// 1a. The setting user is a super admin
	// OR
	// 1. The group being given/removed does not have administrator access,
	// 2. The setting user is an administrator and the target user is not an administrator or super admin.

	////////////////////////////////////////////////////////
	// 1a. If the setting user is super admin
	// Get setting user permissions
	setterPerms, err := s.authorizer.GetPermissions(ctx, domain.AuthScope{Type: domain.AuthObjRefractor}, groupctx.SetterUserID)
	if err != nil {
		s.logger.Error("Could not get setter perms", zap.Error(err))
		return false, err
	}

	if setterPerms.CheckFlag(perms.GetFlag(perms.FlagSuperAdmin)) {
		s.logger.Info("User was granted access to add role to target user",
			zap.String("Setter User ID", groupctx.SetterUserID),
			zap.String("Target User ID", groupctx.TargetUserID),
			zap.Int64("Target Group ID", groupctx.GroupID),
			zap.String("Reason", "Setter was a super admin"),
		)
		return true, nil
	}

	//////////////////////////////////////////////////////////////////
	// 1. The group being given does not have administrator access
	addGroup, err := s.repo.GetByID(ctx, groupctx.GroupID)
	if err != nil {
		s.logger.Error("Could not get target group", zap.Error(err))
		return false, err
	}

	groupPerms, err := bitperms.FromString(addGroup.Permissions)
	if err != nil {
		s.logger.Error("Could not parse target group permissions", zap.Error(err))
		return false, err
	}

	if groupPerms.CheckFlag(perms.GetFlag(perms.FlagAdministrator)) {
		s.logGroupSetDenyMsg(groupctx, "add", "The target group has administrator access and the setting user is not a super admin")
		return false, nil
	}

	///////////////////////////////////////////////////////////////////////
	// 2. The setting user is an administrator and the target user is not an administrator or super admin.
	targetPerms, err := s.authorizer.GetPermissions(ctx, domain.AuthScope{Type: domain.AuthObjRefractor}, groupctx.TargetUserID)
	if err != nil {
		s.logger.Error("Could not get setter perms", zap.Error(err))
		return false, err
	}

	setterIsAdmin := setterPerms.CheckFlag(perms.GetFlag(perms.FlagAdministrator))
	targetIsAdmin := targetPerms.CheckFlag(perms.GetFlag(perms.FlagAdministrator))
	targetIsSuperAdmin := targetPerms.CheckFlag(perms.GetFlag(perms.FlagSuperAdmin))

	if !setterIsAdmin {
		s.logGroupSetDenyMsg(groupctx, "add", "The setter user is not an administrator")
		return false, nil
	} else if setterIsAdmin && targetIsAdmin {
		s.logGroupSetDenyMsg(groupctx, "add", "The target user is an administrator and the setter is not a super admin")
		return false, nil
	} else if setterIsAdmin && targetIsSuperAdmin {
		s.logGroupSetDenyMsg(groupctx, "add", "The target user is a super admin and the setter is not")
		return false, nil
	}

	s.logger.Info("User was granted access to add role to target user",
		zap.String("Setter User ID", groupctx.SetterUserID),
		zap.String("Target User ID", groupctx.TargetUserID),
		zap.Int64("Target Group ID", groupctx.GroupID),
	)

	return true, nil
}

func (s *groupService) RemoveUserGroup(c context.Context, groupctx domain.GroupSetContext) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	canSetGroup, err := s.canSetGroup(ctx, groupctx)
	if err != nil {
		return err
	}

	if !canSetGroup {
		return domain.NewHTTPError(nil, http.StatusUnauthorized,
			"You do not have permission to remove a group from that user.")
	}

	return s.repo.RemoveUserGroup(ctx, groupctx.TargetUserID, groupctx.GroupID)
}

// logGroupSetDenyMsg is a helper function to reduce repetition of logging group add/remove permission deny messages.
func (s *groupService) logGroupSetDenyMsg(groupctx domain.GroupSetContext, roleAction string, reason string) {
	s.logger.Info("User was denied access to "+roleAction+" role to target user",
		zap.String("Setter User ID", groupctx.SetterUserID),
		zap.String("Target User ID", groupctx.TargetUserID),
		zap.Int64("Target Group ID", groupctx.GroupID),
		zap.String("Reason", reason),
	)
}
