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
	"time"
)

type Group struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Color       int       `json:"color"`
	Position    int       `json:"position"`
	Permissions string    `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
}

type GroupRepo interface {
	Store(ctx context.Context, group *Group) error
	GetAll(ctx context.Context) ([]*Group, error)
	GetByID(ctx context.Context, id int64) (*Group, error)
	GetUserGroups(ctx context.Context, userID string) ([]*Group, error)
}

type GroupService interface {
	Store(c context.Context, group *Group) error
	GetAll(c context.Context) ([]*Group, error)
	GetByID(c context.Context, id int64) (*Group, error)
}
