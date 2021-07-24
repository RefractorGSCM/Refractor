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
	"database/sql"
	"sort"
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

type DBGroup struct {
	ID          int64
	Name        string
	Color       int
	Position    int
	Permissions string
	CreatedAt   sql.NullTime
	ModifiedAt  sql.NullTime
}

// GroupSlice is an alias for []*Group. It exists so that we can attach helper functions to it such as sorters.
type GroupSlice []*Group

// SortByPosition sorts the GroupSlice by the position field in ascending order.
func (gs GroupSlice) SortByPosition() []*Group {
	sort.Slice(gs, func(i, j int) bool {
		return gs[i].Position < gs[j].Position
	})

	return gs
}

func (dbg DBGroup) Group() *Group {
	g := &Group{
		ID:          dbg.ID,
		Name:        dbg.Name,
		Color:       dbg.Color,
		Position:    dbg.Position,
		Permissions: dbg.Permissions,
	}

	if dbg.CreatedAt.Valid {
		g.CreatedAt = dbg.CreatedAt.Time
	}

	if dbg.ModifiedAt.Valid {
		g.ModifiedAt = dbg.ModifiedAt.Time
	}

	return g
}

type Overrides struct {
	AllowOverrides string `json:"allow_overrides"`
	DenyOverrides  string `json:"deny_overrides"`
}

type GroupReorderInfo struct {
	GroupID int64 `json:"id"`
	NewPos  int64 `json:"new_pos"`
}

// GroupRepo is an interface defining the behaviour required to manage permissions.
type GroupRepo interface {
	Store(ctx context.Context, group *Group) error
	GetAll(ctx context.Context) ([]*Group, error)
	GetByID(ctx context.Context, id int64) (*Group, error)
	GetUserGroups(ctx context.Context, userID string) ([]*Group, error)
	GetUserOverrides(ctx context.Context, userID string) (*Overrides, error)
	SetUserOverrides(ctx context.Context, userID string, overrides *Overrides) error
	GetBaseGroup(ctx context.Context) (*Group, error)
	SetBaseGroup(ctx context.Context, group *Group) error
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, id int64, args UpdateArgs) (*Group, error)
	Reorder(ctx context.Context, newPositions []*GroupReorderInfo) error
	AddUserGroup(ctx context.Context, userID string, groupID int64) error
	RemoveUserGroup(ctx context.Context, userID string, groupID int64) error
}

type GroupService interface {
	Store(c context.Context, group *Group) error
	GetAll(c context.Context) ([]*Group, error)
	GetByID(c context.Context, id int64) (*Group, error)
	Delete(c context.Context, id int64) error
	Update(c context.Context, id int64, args UpdateArgs) (*Group, error)
	UpdateBase(c context.Context, args UpdateArgs) (*Group, error)
	Reorder(c context.Context, newPositions []*GroupReorderInfo) error
}
