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
	"database/sql"
	"time"
)

type Player struct {
	PlayerID      string    `json:"id"`
	Platform      string    `json:"platform"`
	CurrentName   string    `json:"current_name"`
	PreviousNames []string  `json:"previous_names"`
	Watched       bool      `json:"watched"`
	LastSeen      time.Time `json:"last_seen"`
	CreatedAt     time.Time `json:"created_at"`
	ModifiedAt    time.Time `json:"modified_at"`
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

type PlayerMeta struct {
}
