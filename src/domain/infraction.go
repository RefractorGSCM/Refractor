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
	"github.com/guregu/null"
)

type InfractionType interface {
	Name() string
	AllowedUpdateFields() []string
}

const (
	InfractionTypeWarning = "WARNING"
	InfractionTypeMute    = "MUTE"
	InfractionTypeKick    = "KICK"
	InfractionTypeBan     = "BAN"
)

type Infraction struct {
	InfractionID int64       `json:"id"`
	PlayerID     string      `json:"player_id"`
	Platform     string      `json:"platform"`
	UserID       null.String `json:"user_id"` // UserID is the ID of the user who created this infraction record
	ServerID     int64       `json:"server_id"`
	Type         string      `json:"type"`
	Reason       null.String `json:"reason"`
	Duration     null.Int    `json:"duration"`
	SystemAction bool        `json:"system_action"`
	CreatedAt    null.Time   `json:"created_at"`
	ModifiedAt   null.Time   `json:"modified_at"`
	IssuerName   string      `json:"issuer_name,omitempty"` // IssuerName is not a DB field. It does not get scanned. It is populated manually.
	PlayerName   string      `json:"player_name,omitempty"` // PlayerName is not a DB field. It does not get scanned. It is populated manually.
}

type InfractionRepo interface {
	Store(ctx context.Context, infraction *Infraction) (*Infraction, error)
	GetByID(ctx context.Context, id int64) (*Infraction, error)
	Update(ctx context.Context, id int64, args UpdateArgs) (*Infraction, error)
	Delete(ctx context.Context, id int64) error
	GetByPlayer(ctx context.Context, playerID, platform string) ([]*Infraction, error)
	Search(ctx context.Context, args FindArgs, limit, offset int) (int, []*Infraction, error)
	GetLinkedChatMessages(ctx context.Context, id int64) ([]*ChatMessage, error)
	LinkChatMessages(ctx context.Context, id int64, messageIDs ...int64) error
	UnlinkChatMessages(ctx context.Context, id int64, messageIDs ...int64) error
	PlayerIsBanned(ctx context.Context, platform, playerID string) (bool, int64, error)
}

type InfractionService interface {
	Store(c context.Context, infraction *Infraction, attachments []*Attachment, linkedMessages []int64) (*Infraction, error)
	GetByID(c context.Context, id int64) (*Infraction, error)
	Update(c context.Context, id int64, args UpdateArgs) (*Infraction, error)
	Delete(c context.Context, id int64) error
	GetByPlayer(c context.Context, playerID, platform string) ([]*Infraction, error)
	GetLinkedChatMessages(c context.Context, id int64) ([]*ChatMessage, error)
	LinkChatMessages(c context.Context, id int64, messageIDs ...int64) error
	UnlinkChatMessages(c context.Context, id int64, messageIDs ...int64) error
}
