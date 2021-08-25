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
	"Refractor/pkg/bitperms"
	"Refractor/pkg/perms"
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
