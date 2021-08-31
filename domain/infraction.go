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

const (
	InfractionTypeWarning = "WARNING"
	InfractionTypeMute    = "MUTE"
	InfractionTypeKick    = "KICK"
	InfractionTypeBan     = "BAN"
)

type Infraction struct {
	InfractionID int64     `json:"id"`
	PlayerID     string    `json:"player_id"`
	Platform     string    `json:"platform"`
	UserID       string    `json:"user_id"` // UserID is the ID of the user who created this infraction record
	ServerID     int64     `json:"server_id"`
	Type         string    `json:"type"`
	Reason       string    `json:"reason"`
	Duration     int       `json:"duration"`
	SystemAction bool      `json:"system_action"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}

type DBInfraction struct {
	InfractionID int64
	PlayerID     string
	Platform     string
	UserID       sql.NullString
	ServerID     int64
	Type         string
	Reason       sql.NullString
	Duration     sql.NullInt32
	SystemAction bool
	CreatedAt    sql.NullTime
	ModifiedAt   sql.NullTime
}

func (dbi *DBInfraction) Infraction() *Infraction {
	infraction := &Infraction{
		InfractionID: dbi.InfractionID,
		PlayerID:     dbi.PlayerID,
		UserID:       dbi.Reason.String,
		ServerID:     dbi.ServerID,
		Reason:       dbi.Reason.String,
		Duration:     int(dbi.Duration.Int32),
		Type:         dbi.Type,
		SystemAction: dbi.SystemAction,
	}

	if dbi.CreatedAt.Valid {
		infraction.CreatedAt = dbi.CreatedAt.Time
	}

	if dbi.ModifiedAt.Valid {
		infraction.ModifiedAt = dbi.ModifiedAt.Time
	}

	return infraction
}

type InfractionRepo interface {
	Store(ctx context.Context, infraction *DBInfraction) (*Infraction, error)
	GetByID(ctx context.Context, id int64) (*Infraction, error)
	Update(ctx context.Context, id int64, args UpdateArgs) (*Infraction, error)
	Delete(ctx context.Context, id int64) error
}

type InfractionService interface {
	Store(c context.Context, infraction *Infraction) error
}
