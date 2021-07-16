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
	"strings"
)

// perms is a package which provides supporting functionality for Refractor's binary permission system.
// This package contains the flag definitions. Each flag is a power of two. These powers are retrieved using
// the bitperms.GetFlag() helper function which automatically does the shifting for us.

const (
	FlagAdministrator = FlagName("FLAG_ADMINISTRATOR")
	FlagViewServers   = FlagName("FLAG_VIEW_SERVERS")
)

type FlagName string

type flag struct {
	name        FlagName
	description string
}

var flags = map[FlagName]*big.Int{}
var descriptions = map[FlagName]string{}
var defaultPermissions *bitperms.Permissions

func init() {
	// Register permission flags
	//////////////////////////////////////////////////////////
	// !! DO NOT CHANGE THE ORDER OF THE FLAG REGISTRATIONS !!
	//////////////////////////////////////////////////////////
	// If you need to add new flags, add them to the bottom
	// of the list to avoid changing offsets. If the order changes, it will be break permissions for existing
	// installations of Refractor!
	registerFlags([]flag{
		{
			name: FlagAdministrator,
			description: `Grants full access to Refractor. Administrator is required to be able to add, edit and delete
						  servers as well as modify admin level settings. Only give this permission to people who
						  absolutely need it.`,
		},
		{
			name:        FlagViewServers,
			description: "Allows viewing of servers",
		},
		// ADD NEW FLAGS HERE. Do not touch any of the above flags!
	})

	// Create default permissions value
	defaultPermissions = bitperms.NewPermissionBuilder().
		AddFlag(GetFlag(FlagViewServers)).
		GetPermission()
}

func registerFlags(newFlags []flag) {
	var i uint = 0

	for _, flag := range newFlags {
		next := bitperms.GetFlag(i)
		i++

		flags[flag.name] = next
		descriptions[flag.name] = flag.description
	}
}

// GetFlag returns a flag's integer value.
func GetFlag(flag FlagName) *big.Int {
	return flags[flag]
}

// GetDescription returns a flag's human readable description with newline and tab characters stripped off.
func GetDescription(flag FlagName) string {
	desc := descriptions[flag]
	desc = strings.Replace(desc, "\n", "", -1)
	desc = strings.Replace(desc, "\t", "", -1)

	return desc
}
