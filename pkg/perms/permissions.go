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

package perms

import (
	"Refractor/pkg/bitperms"
	"math/big"
	"regexp"
)

// perms is a package which provides supporting functionality for Refractor's binary Permission system.
// This package contains the Permission definitions. Each Permission is a power of two. These powers are retrieved using
// the bitperms.GetFlag() helper function which automatically does the shifting for us.

const (
	FlagSuperAdmin            = FlagName("FLAG_SUPER_ADMIN")
	FlagAdministrator         = FlagName("FLAG_ADMINISTRATOR")
	FlagViewServers           = FlagName("FLAG_VIEW_SERVERS")
	FlagViewPlayerRecords     = FlagName("FLAG_VIEW_PLAYER_RECORDS")
	FlagViewInfractionRecords = FlagName("FLAG_VIEW_INFRACTION_RECORDS")
	FlagViewChatRecords       = FlagName("FLAG_VIEW_CHAT_RECORDS")
)

type FlagName string

type Permission struct {
	ID          int
	Name        FlagName
	DisplayName string
	Description string
	Flag        *big.Int
}

var permissions = map[FlagName]*Permission{}
var permissionsArr []*Permission
var defaultPermissions *bitperms.Permissions

func init() {
	// Register Permission permissions
	/////////////////////////////////////////////////////
	// !! DO NOT CHANGE THE ORDER OF THE REGISTRATIONS !!
	/////////////////////////////////////////////////////
	// If you need to add new permissions, add them to the bottom
	// of the list to avoid changing offsets. If the order changes, it will be break permissions for existing
	// installations of Refractor!
	registerPermissions([]Permission{
		{
			Name:        FlagSuperAdmin,
			DisplayName: "Super Admin",
			Description: `Grants full access to Refractor including management of admin users, roles, etc. This should
						  NEVER be granted to anybody except for the initial user account in Refractor. No more than one
						  user should have this permission at a time. Seriously, never manually set this permission!`,
		},
		{
			Name:        FlagAdministrator,
			DisplayName: "Administrator",
			Description: `Grants full access to Refractor. Administrator is required to be able to add, edit and delete
						  servers as well as modify admin level settings. Admins can not create or edit groups. Only give
						  this Permission to people who absolutely need it.`,
		},
		{
			Name:        FlagViewServers,
			DisplayName: "View servers",
			Description: "Allows viewing of servers.",
		},
		{
			Name:        FlagViewPlayerRecords,
			DisplayName: "View player records",
			Description: `Allows viewing of player records. This permissions can be overridden on the server level to
						  allow or deny accessing player records for individual servers.`,
		},
		{
			Name:        FlagViewInfractionRecords,
			DisplayName: "View infraction records",
			Description: `Allows viewing of infraction records. This permissions can be overridden on the server level to
						  allow or deny accessing infraction records for individual servers.`,
		},
		{
			Name:        FlagViewChatRecords,
			DisplayName: "View chat records",
			Description: `Allows viewing of chat records. This permissions can be overridden on the server level to
						  allow or deny accessing chat records for individual servers.`,
		},
		// ADD NEW FLAGS HERE. Do not touch any of the above permissions!
	})

	// Create default permissions value
	defaultPermissions = bitperms.NewPermissionBuilder().
		AddFlag(GetFlag(FlagViewServers)).
		AddFlag(GetFlag(FlagViewPlayerRecords)).
		GetPermission()
}

func registerPermissions(newPerms []Permission) {
	var i uint = 0

	for _, perm := range newPerms {
		next := bitperms.GetFlag(i)
		i++

		newPermission := &Permission{
			ID:          int(i),
			Name:        perm.Name,
			DisplayName: perm.DisplayName,
			Description: perm.Description,
			Flag:        next,
		}

		permissions[perm.Name] = newPermission
		permissionsArr = append(permissionsArr, newPermission)
	}
}

// GetFlag returns a Permission's integer value.
func GetFlag(flag FlagName) *big.Int {
	return permissions[flag].Flag
}

func GetAll() []*Permission {
	return permissionsArr
}

var whitespacePattern = regexp.MustCompile("\\s+")

// GetDescription returns a Permission's human readable Description with newline and tab characters stripped off.
func GetDescription(flag FlagName) string {
	desc := permissions[flag].Description
	desc = whitespacePattern.ReplaceAllString(desc, " ")

	return desc
}

func GetDefaultPermissions() *bitperms.Permissions {
	return defaultPermissions
}
