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
	FlagSuperAdmin    = FlagName("FLAG_SUPER_ADMIN")
	FlagAdministrator = FlagName("FLAG_ADMINISTRATOR")
	FlagViewServers   = FlagName("FLAG_VIEW_SERVERS")
)

type FlagName string

type Permission struct {
	Name        FlagName
	Description string
	Flag        *big.Int
}

var permissions = map[FlagName]*Permission{}
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
			Name: FlagSuperAdmin,
			Description: `Grants full access to Refractor including management of admin users, roles, etc. This should
						  NEVER be granted to anybody except for the initial user account in Refractor. No more than one
						  user should have this permission at a time. Seriously, never manually set this permission!`,
		},
		{
			Name: FlagAdministrator,
			Description: `Grants full access to Refractor. Administrator is required to be able to add, edit and delete
						  servers as well as modify admin level settings. Only give this Permission to people who
						  absolutely need it.`,
		},
		{
			Name:        FlagViewServers,
			Description: "Allows viewing of servers",
		},
		// ADD NEW FLAGS HERE. Do not touch any of the above permissions!
	})

	// Create default permissions value
	defaultPermissions = bitperms.NewPermissionBuilder().
		AddFlag(GetFlag(FlagViewServers)).
		GetPermission()
}

func registerPermissions(newPerms []Permission) {
	var i uint = 0

	for _, perm := range newPerms {
		next := bitperms.GetFlag(i)
		i++

		permissions[perm.Name] = &Permission{
			Name:        perm.Name,
			Description: perm.Description,
			Flag:        next,
		}
	}
}

// GetFlag returns a Permission's integer value.
func GetFlag(flag FlagName) *big.Int {
	return permissions[flag].Flag
}

func GetAll() []*Permission {
	var all []*Permission

	for _, val := range permissions {
		all = append(all, val)
	}

	return all
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
