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

func (a *authorizer) hasPermissionRefractor(ctx context.Context, userID string, checkAuth domain.AuthChecker) (bool, error) {
	const op = opTag + "hasPermissionRefractor"

	computedPerms, err := a.computePermissionsRefractor(ctx, userID)
	if err != nil {
		return false, errors.Wrap(err, op)
	}

	return checkAuth(computedPerms)
}

func (a *authorizer) computePermissionsRefractor(ctx context.Context, userID string) (*bitperms.Permissions, error) {
	const op = opTag + "computePermissionsRefractor"

	// Permissions are computed in the following order:
	// 1. Base permissions given to everyone (default group) at the application level
	// 2. Permissions allowed to a user by their groups at the application level
	// 3. NewUser-specific overrides that deny permissions at the application level
	// 4. NewUser-specific overrides that allow permissions at the application level
	// The final calculated result is the user's fully computed permissions.

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

	// 3a Get user overrides
	userOverrides, err := a.groupRepo.GetUserOverrides(ctx, userID)
	if err != nil && errors.Cause(err) != domain.ErrNotFound {
		a.logger.Error("Could not get user overrides", zap.String("UserID", userID), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	// If the user has overrides set, then compute them
	if userOverrides != nil {
		// 3. Compute user deny overrides
		basePermissions, err = basePermissions.ComputeDenyOverrides(userOverrides.DenyOverrides)
		if err != nil {
			a.logger.Error("Could not compute user deny overrides", zap.Error(err))
			return nil, errors.Wrap(err, op)
		}

		// 4. Compute user allow overrides
		basePermissions, err = basePermissions.ComputeAllowOverrides(userOverrides.AllowOverrides)
		if err != nil {
			a.logger.Error("Could not compute user allow overrides", zap.Error(err))
			return nil, errors.Wrap(err, op)
		}
	}

	return basePermissions, nil
}
