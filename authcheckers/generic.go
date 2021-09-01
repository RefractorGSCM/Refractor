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

package authcheckers

import (
	"Refractor/domain"
	"Refractor/pkg/bitperms"
	"Refractor/pkg/perms"
	"github.com/pkg/errors"
)

func DenyAll(permissions *bitperms.Permissions) (bool, error) {
	return false, nil
}

func AllowAll(permissions *bitperms.Permissions) (bool, error) {
	return true, nil
}

func RequireAdmin(permissions *bitperms.Permissions) (bool, error) {
	if permissions.CheckFlag(perms.GetFlag(perms.FlagAdministrator)) {
		return true, nil
	}

	return false, nil
}

func HasPermission(flagName perms.FlagName, adminBypass bool) domain.AuthChecker {
	return func(permissions *bitperms.Permissions) (bool, error) {
		flag := perms.GetFlag(flagName)
		if flag == nil {
			return false, errors.New("invalid flag name")
		}

		if permissions.CheckFlag(flag) {
			return true, nil
		}

		// Super admins bypass all permission checks
		if permissions.CheckFlag(perms.GetFlag(perms.FlagSuperAdmin)) {
			return true, nil
		}

		// If adminBypass is enabled and the specified flag is not set, check if the user is admin
		if permissions.CheckFlag(perms.GetFlag(perms.FlagAdministrator)) {
			return true, nil
		}

		return false, nil
	}
}

func CanViewServer(permissions *bitperms.Permissions) (bool, error) {
	if permissions.CheckFlag(perms.GetFlag(perms.FlagViewServers)) {
		return true, nil
	}

	if permissions.CheckFlag(perms.GetFlag(perms.FlagAdministrator)) {
		return true, nil
	}

	if permissions.CheckFlag(perms.GetFlag(perms.FlagSuperAdmin)) {
		return true, nil
	}

	return false, nil
}

func CanViewPlayerRecords(permissions *bitperms.Permissions) (bool, error) {
	if permissions.CheckFlag(perms.GetFlag(perms.FlagViewPlayerRecords)) {
		return true, nil
	}

	if permissions.CheckFlag(perms.GetFlag(perms.FlagAdministrator)) {
		return true, nil
	}

	return false, nil
}
