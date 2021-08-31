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

package params

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"math"
	"strings"
)

type CreateWarningParams struct {
	PlayerID int64  `json:"player_id" form:"player_id"`
	Platform string `json:"platform" form:"platform"`
	ServerID int64  `json:"server_id" form:"server_id"`
	Reason   string `json:"reason" form:"reason"`
}

func (body CreateWarningParams) Validate() error {
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		validation.Field(&body.PlayerID, validation.Required, validation.Length(1, 80)),
		validation.Field(&body.Platform, validation.Required, validation.Length(1, 128)),
		validation.Field(&body.ServerID, validation.Required, is.Int, validation.Min(1), validation.Max(math.MaxInt32)),
		validation.Field(&body.Reason, validation.Required, validation.Length(1, 1024)))
}

type CreateMuteParams struct {
	PlayerID int64  `json:"player_id" form:"player_id"`
	Platform string `json:"platform" form:"platform"`
	ServerID int64  `json:"server_id" form:"server_id"`
	Reason   string `json:"reason" form:"reason"`
	Duration int    `json:"duration" form:"duration"`
}

func (body CreateMuteParams) Validate() error {
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		validation.Field(&body.PlayerID, validation.Required, validation.Length(1, 80)),
		validation.Field(&body.Platform, validation.Required, validation.Length(1, 128)),
		validation.Field(&body.ServerID, validation.Required, is.Int, validation.Min(1), validation.Max(math.MaxInt32)),
		validation.Field(&body.Reason, validation.Required, validation.Length(1, 1024)),
		validation.Field(&body.Duration, validation.Required, is.Int, validation.Min(0), validation.Max(math.MaxInt32)))
}

type CreateKickParams struct {
	PlayerID int64  `json:"player_id" form:"player_id"`
	Platform string `json:"platform" form:"platform"`
	ServerID int64  `json:"server_id" form:"server_id"`
	Reason   string `json:"reason" form:"reason"`
}

func (body CreateKickParams) Validate() error {
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		validation.Field(&body.PlayerID, validation.Required, validation.Length(1, 80)),
		validation.Field(&body.Platform, validation.Required, validation.Length(1, 128)),
		validation.Field(&body.ServerID, validation.Required, is.Int, validation.Min(1), validation.Max(math.MaxInt32)),
		validation.Field(&body.Reason, validation.Required, validation.Length(1, 1024)))
}
