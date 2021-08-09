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
	"time"
)

type Server struct {
	ID           int64     `json:"id"`
	Game         string    `json:"game"`
	Name         string    `json:"string"`
	Address      string    `json:"address"`
	RCONPort     string    `json:"rcon_port"`
	RCONPassword string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type DBServer struct {
	ID           int64
	Game         string
	Name         string
	Address      string
	RCONPort     string
	RCONPassword string
	CreatedAt    sql.NullTime
	ModifiedAt   sql.NullTime
}

func (dbs DBServer) Server() *Server {
	s := &Server{
		ID:           dbs.ID,
		Game:         dbs.Game,
		Name:         dbs.Name,
		Address:      dbs.Address,
		RCONPort:     dbs.RCONPort,
		RCONPassword: dbs.RCONPassword,
	}

	if dbs.CreatedAt.Valid {
		s.CreatedAt = dbs.CreatedAt.Time
	}

	if dbs.ModifiedAt.Valid {
		s.ModifiedAt = dbs.ModifiedAt.Time
	}

	return s
}

type ServerRepo interface {
	Store(ctx context.Context, server *Server) error
	GetByID(ctx context.Context, id int64) (*Server, error)
	GetAll(c context.Context) ([]*Server, error)
}

type ServerService interface {
	Store(c context.Context, server *Server) error
	GetByID(c context.Context, id int64) (*Server, error)
	GetAll(c context.Context) ([]*Server, error)
}
