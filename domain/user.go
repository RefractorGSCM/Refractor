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
	"context"
)

type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Permissions string    `json:"permissions"`
	Groups      []*Group  `json:"groups"`
	UserMeta    *UserMeta `json:"meta"`
}

type UserMeta struct {
	ID              string `json:"id"`
	InitialUsername string `json:"initial_username"`
	Username        string `json:"username"`
	Deactivated     bool   `json:"deactivated"`
}

// UserMetaRepo is the interface to handle the storing of UserMeta data. This is NOT an auth repository and only contains
// relevant metadata for Refractor. No user identities are stored in a UserMetaRepo!
type UserMetaRepo interface {
	Store(ctx context.Context, userInfo *UserMeta) error
	GetByID(ctx context.Context, userID string) (*UserMeta, error)
	Update(ctx context.Context, userID string, args UpdateArgs) (*UserMeta, error)
	IsDeactivated(ctx context.Context, userID string) (bool, error)
}

type UserService interface {
	GetAllUsers(c context.Context) ([]*User, error)
	GetByID(c context.Context, userID string) (*User, error)
	DeactivateUser(c context.Context, userID string) error
	ReactivateUser(c context.Context, userID string) error
}
