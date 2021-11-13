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
	"Refractor/pkg/broadcast"
	"context"
	"database/sql"
	"time"
)

type Player struct {
	PlayerID                     string    `json:"id"`
	Platform                     string    `json:"platform"`
	CurrentName                  string    `json:"name"`
	PreviousNames                []string  `json:"previous_names"`
	Watched                      bool      `json:"watched"`
	LastSeen                     time.Time `json:"last_seen"`
	CreatedAt                    time.Time `json:"created_at"`
	ModifiedAt                   time.Time `json:"modified_at"`
	InfractionCount              int       `json:"infraction_count,omitempty"`                // not a db field. must be populated manually
	InfractionCountSinceTimespan int       `json:"infraction_count_since_timespan,omitempty"` // not a db field. must be populated manually
}

type DBPlayer struct {
	PlayerID      string
	Platform      string
	CurrentName   string
	PreviousNames []string
	Watched       bool
	LastSeen      sql.NullTime
	CreatedAt     sql.NullTime
	ModifiedAt    sql.NullTime
}

func (dbp DBPlayer) Player() *Player {
	player := &Player{
		PlayerID:      dbp.PlayerID,
		Platform:      dbp.Platform,
		CurrentName:   dbp.CurrentName,
		PreviousNames: dbp.PreviousNames,
		Watched:       dbp.Watched,
	}

	if dbp.LastSeen.Valid {
		player.LastSeen = dbp.LastSeen.Time
	}

	if dbp.CreatedAt.Valid {
		player.CreatedAt = dbp.CreatedAt.Time
	}

	if dbp.ModifiedAt.Valid {
		player.ModifiedAt = dbp.ModifiedAt.Time
	}

	return player
}

type PlayerRepo interface {
	Store(ctx context.Context, player *Player) error
	GetByID(ctx context.Context, platform, id string) (*Player, error)
	Exists(ctx context.Context, args FindArgs) (bool, error)
	Update(ctx context.Context, platform, id string, args UpdateArgs) (*Player, error)
	SearchByName(ctx context.Context, name string, limit, offset int) (int, []*Player, error)
}

type PlayerNameRepo interface {
	Store(ctx context.Context, id, platform, name string) error
	GetNames(ctx context.Context, id, platform string) (string, []string, error)
	UpdateName(ctx context.Context, player *Player, newName string) error
}

type PlayerService interface {
	HandlePlayerJoin(fields broadcast.Fields, serverID int64, game Game)
	HandlePlayerQuit(fields broadcast.Fields, serverID int64, game Game)
	GetPlayer(c context.Context, id, platform string) (*Player, error)
}
