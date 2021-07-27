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
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Permissions string   `json:"permissions"`
	Groups      []*Group `json:"groups"`
}

type UserInfo struct {
	ID              string `json:"id"`
	InitialUsername string `json:"initial_username"`
	Username        string `json:"username"`
	Deactivated     bool   `json:"deactivated"`
}

// UserRepo is the interface to handle the storing of UserInfo data. This is NOT an auth repository and only contains
// relevant metadata for Refractor. No user identities are stored in a UserRepo!
type UserRepo interface {
	Store(ctx context.Context, userInfo *UserInfo) error
	GetByID(ctx context.Context, userID string) (*UserInfo, error)
	SetUsername(ctx context.Context, username string) error
}

type UserService interface {
	GetAllUsers(c context.Context) ([]*User, error)
}
