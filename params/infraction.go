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

func infractionReasonRules(fieldPtr interface{}) *validation.FieldRules {
	return validation.Field(fieldPtr, validation.Required, validation.Length(1, 1024))
}

func infractionDurationRules(fieldPtr interface{}) *validation.FieldRules {
	return validation.Field(fieldPtr, validation.Required, is.Int, validation.Min(0), validation.Max(math.MaxInt32))
}

type CreateWarningParams struct {
	PlayerID string `json:"player_id" form:"player_id"`
	Platform string `json:"platform" form:"platform"`
	Reason   string `json:"reason" form:"reason"`
}

func (body CreateWarningParams) Validate() error {
	body.PlayerID = strings.TrimSpace(body.Reason)
	body.Platform = strings.TrimSpace(body.Platform)
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		playerIDRules(&body.PlayerID),
		platformRules(&body.Platform),
		infractionReasonRules(&body.Reason),
	)
}

type CreateMuteParams struct {
	PlayerID string `json:"player_id" form:"player_id"`
	Platform string `json:"platform" form:"platform"`
	Reason   string `json:"reason" form:"reason"`
	Duration int    `json:"duration" form:"duration"`
}

func (body CreateMuteParams) Validate() error {
	body.PlayerID = strings.TrimSpace(body.Reason)
	body.Platform = strings.TrimSpace(body.Platform)
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		playerIDRules(&body.PlayerID),
		platformRules(&body.Platform),
		infractionReasonRules(&body.Reason),
		infractionDurationRules(&body.Duration),
	)
}

type CreateKickParams struct {
	PlayerID string `json:"player_id" form:"player_id"`
	Platform string `json:"platform" form:"platform"`
	Reason   string `json:"reason" form:"reason"`
}

func (body CreateKickParams) Validate() error {
	body.PlayerID = strings.TrimSpace(body.Reason)
	body.Platform = strings.TrimSpace(body.Platform)
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		playerIDRules(&body.PlayerID),
		platformRules(&body.Platform),
		infractionReasonRules(&body.Reason),
	)
}

type CreateBanParams struct {
	PlayerID string `json:"player_id" form:"player_id"`
	Platform string `json:"platform" form:"platform"`
	Reason   string `json:"reason" form:"reason"`
	Duration int    `json:"duration" form:"duration"`
}

func (body CreateBanParams) Validate() error {
	body.PlayerID = strings.TrimSpace(body.Reason)
	body.Platform = strings.TrimSpace(body.Platform)
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		playerIDRules(&body.PlayerID),
		platformRules(&body.Platform),
		infractionReasonRules(&body.Reason),
		infractionDurationRules(&body.Duration),
	)
}
