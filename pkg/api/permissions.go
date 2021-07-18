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

package api

import (
	"Refractor/domain"
	"Refractor/pkg/bitperms"
	"Refractor/pkg/perms"
	"context"
)

// CheckPermissions is a wrapper function which provides automatic checking of if a user is a superadmin.
func CheckPermissions(ctx context.Context, a domain.Authorizer, scope domain.AuthScope, userID string, authChecker domain.AuthChecker) (bool, error) {
	hasPermission, err := a.HasPermission(ctx, scope, userID, func(permissions *bitperms.Permissions) (bool, error) {
		if permissions.CheckFlag(perms.GetFlag(perms.FlagSuperAdmin)) {
			return true, nil
		}

		return authChecker(permissions)
	})
	if err != nil {
		return false, err
	}

	return hasPermission, nil
}
