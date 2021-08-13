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

package authorizer

import (
	"Refractor/domain"
	"Refractor/pkg/bitperms"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (a *authorizer) hasPermissionServer(ctx context.Context, userID string, serverID int64, comparator domain.AuthChecker) (bool, error) {
	return false, nil
}

func (a *authorizer) computePermissionsServer(ctx context.Context, userID string, serverID int64) (*bitperms.Permissions, error) {
	const op = opTag + "computePermissionsServer"

	// Permissions are computed in the following order:
	// 1. Base permissions given to everyone (default group) at the application level
	// 2. Permissions allowed to a user by their groups at the application level
	// 3. Server Group overrides that deny permissions at the server level
	// 4. Server Group overrides that allow permissions at the server level
	// 5. User-specific overrides that deny permissions at the application level
	// 6. User-specific overrides that allow permissions at the application level

	// 1. Compute base permissions
	groupEveryone, err := a.groupRepo.GetBaseGroup(ctx)
	if err != nil {
		a.logger.Error("Could not get base group", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	basePermissions, err := bitperms.FromString(groupEveryone.Permissions)
	if err != nil {
		a.logger.Error("Could not parse base group permissions", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	// 2. Compute permissions granted by the user's groups
	userGroups, err := a.groupRepo.GetUserGroups(ctx, userID)
	if err != nil && errors.Cause(err) != domain.ErrNotFound {
		a.logger.Error("Could not get user groups", zap.String("UserID", userID), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	for _, group := range userGroups {
		groupPerms, err := bitperms.FromString(group.Permissions)
		if err != nil {
			a.logger.Error("Could not parse permissions", zap.Int64("GroupID", group.ID), zap.Error(err))
			return nil, errors.Wrap(err, op)
		}

		// Or together the base perms and the group's perms to get the combined value
		basePermissions = basePermissions.Or(groupPerms)
	}

	// pre-3. Get group server overrides for all user groups
	var groupOverrides = map[int64]*domain.Overrides{}

	for _, group := range userGroups {
		overrides, err := a.groupRepo.GetServerOverrides(ctx, serverID, group.ID)
		if err != nil {
			a.logger.Error(
				"Could not get group server overrides",
				zap.Int64("Server ID", serverID),
				zap.Int64("Group ID", group.ID),
				zap.Error(err),
			)
			return nil, errors.Wrap(err, op)
		}

		groupOverrides[group.ID] = overrides
	}

	// 3. Compute server group deny overrides
	for _, group := range userGroups {
		// If server overrides exist for this group, compute them
		if groupOverrides[group.ID] != nil {
			overrides := groupOverrides[group.ID]

			basePermissions, err = basePermissions.ComputeDenyOverrides(overrides.DenyOverrides)
			if err != nil {
				a.logger.Error(
					"Could not compute group server deny overrides",
					zap.Int64("Server ID", serverID),
					zap.Int64("Group ID", group.ID),
					zap.Error(err),
				)
				return nil, errors.Wrap(err, op)
			}
		}
	}

	// 4. Compute server group allow overrides
	for _, group := range userGroups {
		// If server overrides exist for this group, compute them
		if groupOverrides[group.ID] != nil {
			overrides := groupOverrides[group.ID]

			basePermissions, err = basePermissions.ComputeAllowOverrides(overrides.AllowOverrides)
			if err != nil {
				a.logger.Error(
					"Could not compute group server allow overrides",
					zap.Int64("Server ID", serverID),
					zap.Int64("Group ID", group.ID),
					zap.Error(err),
				)
				return nil, errors.Wrap(err, op)
			}
		}
	}

	// pre-5. Get user overrides
	userOverrides, err := a.groupRepo.GetUserOverrides(ctx, userID)
	if err != nil && errors.Cause(err) != domain.ErrNotFound {
		a.logger.Error("Could not get user overrides", zap.String("UserID", userID), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	// If the user has overrides set, then compute them
	if userOverrides != nil {
		// 5. Compute user deny overrides
		basePermissions, err = basePermissions.ComputeDenyOverrides(userOverrides.DenyOverrides)
		if err != nil {
			a.logger.Error("Could not compute user deny overrides", zap.Error(err))
			return nil, errors.Wrap(err, op)
		}

		// 6. Compute user allow overrides
		basePermissions, err = basePermissions.ComputeAllowOverrides(userOverrides.AllowOverrides)
		if err != nil {
			a.logger.Error("Could not compute user allow overrides", zap.Error(err))
			return nil, errors.Wrap(err, op)
		}
	}

	return basePermissions, nil
}
