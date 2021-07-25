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

package domain

import (
	"Refractor/pkg/bitperms"
	"context"
	kratos "github.com/ory/kratos-client-go"
)

type Traits struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type AuthUser struct {
	Traits *Traits
	*kratos.Session
}

type AuthRepo interface {
	CreateUser(ctx context.Context, userTraits *Traits) (*AuthUser, error)
	GetUserByID(ctx context.Context, id string) (*AuthUser, error)
	GetAllUsers(ctx context.Context) ([]*AuthUser, error)
	GetRecoveryLink(ctx context.Context, userID string) (string, error)
}

type AuthService interface {
	CreateUser(c context.Context, userTraits *Traits, inviter string) (*AuthUser, error)
}

type AuthObject string

const (
	AuthObjRefractor = AuthObject("REFRACTOR")
	AuthObjServer    = AuthObject("SERVER")
)

type AuthScope struct {
	// Type represents the object type being authenticated against. e.g SERVER
	Type AuthObject

	// ID represents an object identifier. e.g 1 (Server ID)
	ID interface{}

	// Context is a map where data can be added for use in authorization checkers
	Context map[string]interface{}
}

type AuthChecker func(permissions *bitperms.Permissions) (bool, error)

// Authorizer represents an entity which can be used to determine if a user has access to perform an action or not.
type Authorizer interface {
	HasPermission(ctx context.Context, scope AuthScope, userID string, authChecker AuthChecker) (bool, error)
	GetPermissions(ctx context.Context, scope AuthScope, userID string) (*bitperms.Permissions, error)
}
