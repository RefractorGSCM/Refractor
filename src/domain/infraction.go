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
	"github.com/guregu/null"
	"math"
	"time"
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

	// PermanentInfractionValue will be used as a duration to mark an infraction as permanent
	PermanentInfractionValue = -1
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
	Repealed     bool        `json:"repealed"`
	IssuerName   string      `json:"issuer_name,omitempty"` // IssuerName is not a DB field. It does not get scanned. It is populated manually.
	PlayerName   string      `json:"player_name,omitempty"` // PlayerName is not a DB field. It does not get scanned. It is populated manually.
}

func (i *Infraction) IsPermanent() bool {
	return i.Duration.Valid && i.Duration.ValueOrZero() == -1
}

func (i *Infraction) MinutesRemaining() int64 {
	if !i.CreatedAt.Valid {
		return 0
	}

	if !i.Duration.Valid {
		return 0
	}

	expiresAt := i.CreatedAt.ValueOrZero().Add(time.Duration(i.Duration.ValueOrZero() * int64(time.Minute)))
	return int64(math.Round(expiresAt.Sub(time.Now()).Minutes()))
}

type InfractionRepo interface {
	Store(ctx context.Context, infraction *Infraction) (*Infraction, error)
	GetByID(ctx context.Context, id int64) (*Infraction, error)
	Update(ctx context.Context, id int64, args UpdateArgs) (*Infraction, error)
	Delete(ctx context.Context, id int64) error
	GetByPlayer(ctx context.Context, playerID, platform string) ([]*Infraction, error)
	Search(ctx context.Context, args FindArgs, serverIDs []int64, limit, offset int) (int, []*Infraction, error)
	GetLinkedChatMessages(ctx context.Context, id int64) ([]*ChatMessage, error)
	LinkChatMessages(ctx context.Context, id int64, messageIDs ...int64) error
	UnlinkChatMessages(ctx context.Context, id int64, messageIDs ...int64) error

	// GetMostSignificantInfraction should return the non-expired, non-repealed infraction of the
	// specified type with the highest duration.
	GetMostSignificantInfraction(ctx context.Context, infrType, platform, playerID string) (*Infraction, error)
	GetPlayerTotalInfractions(ctx context.Context, platform, playerID string) (int, error)
	GetPlayerInfractionCountSince(ctx context.Context, platform, playerID string, since time.Time) (int, error)
}

type InfractionSubscriber func(infraction *Infraction)

type InfractionService interface {
	Store(c context.Context, infraction *Infraction, attachments []*Attachment, linkedMessages []int64) (*Infraction, error)
	GetByID(c context.Context, id int64) (*Infraction, error)
	Update(c context.Context, id int64, args UpdateArgs) (*Infraction, error)
	SetRepealed(c context.Context, id int64, repealed bool) (*Infraction, error)
	Delete(c context.Context, id int64) error
	GetByPlayer(c context.Context, playerID, platform string) ([]*Infraction, error)
	GetLinkedChatMessages(c context.Context, id int64) ([]*ChatMessage, error)
	LinkChatMessages(c context.Context, id int64, messageIDs ...int64) error
	UnlinkChatMessages(c context.Context, id int64, messageIDs ...int64) error
	GetCurrentBan(c context.Context, platform, playerID string) (*Infraction, error)
	GetCurrentMute(c context.Context, platform, playerID string) (*Infraction, error)
	HandlePlayerJoin(fields broadcast.Fields, serverID int64, game Game)
	HandleModerationAction(fields broadcast.Fields, serverID int64, game Game)
	SubscribeInfractionCreate(sub InfractionSubscriber)
}

const (
	InfractionCommandCreate = "CREATE"
	InfractionCommandUpdate = "UPDATE"
	InfractionCommandDelete = "DELETE"
	InfractionCommandRepeal = "REPEAL"
	InfractionCommandSync   = "SYNC"
)

type InfractionCommands struct {
	Warn []*InfractionCommand `json:"warn"`
	Mute []*InfractionCommand `json:"mute"`
	Kick []*InfractionCommand `json:"kick"`
	Ban  []*InfractionCommand `json:"ban"`
}

type InfractionCommand struct {
	Command string `json:"command"`
	// If RunOnAll is set to true, when triggered the command will be run on all servers with the same game as the
	// server it was issued on. Otherwise, it will only be run on the server it was triggered on.
	RunOnAll bool `json:"run_on_all"`
}

func (ic *InfractionCommands) Map() map[string][]*InfractionCommand {
	return map[string][]*InfractionCommand{
		InfractionTypeWarning: ic.Warn,
		InfractionTypeMute:    ic.Mute,
		InfractionTypeKick:    ic.Kick,
		InfractionTypeBan:     ic.Ban,
	}
}

// Prepare will replace all nil fields with empty arrays for a consistent experience on the frontend
func (ic *InfractionCommands) Prepare() *InfractionCommands {
	if ic.Warn == nil {
		ic.Warn = make([]*InfractionCommand, 0)
	}
	if ic.Mute == nil {
		ic.Mute = make([]*InfractionCommand, 0)
	}
	if ic.Kick == nil {
		ic.Kick = make([]*InfractionCommand, 0)
	}
	if ic.Ban == nil {
		ic.Ban = make([]*InfractionCommand, 0)
	}

	return ic
}

type InfractionPayload interface {
	GetPlayerID() string
	GetPlatform() string
	GetPlayerName() string
	GetType() string
	GetDuration() int64
	GetDurationRemaining() int64
	GetReason() string
	GetUserID() string
}

func (i *Infraction) GetPlayerID() string {
	return i.PlayerID
}

func (i *Infraction) GetPlatform() string {
	return i.Platform
}

func (i *Infraction) GetPlayerName() string {
	return i.PlayerName
}

func (i *Infraction) GetType() string {
	return i.Type
}

func (i *Infraction) GetDuration() int64 {
	return i.Duration.ValueOrZero()
}

func (i *Infraction) GetDurationRemaining() int64 {
	return i.MinutesRemaining()
}

func (i *Infraction) GetReason() string {
	return i.Reason.ValueOrZero()
}

func (i *Infraction) GetUserID() string {
	return i.UserID.ValueOrZero()
}
