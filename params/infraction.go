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
	"Refractor/params/rules"
	validation "github.com/go-ozzo/ozzo-validation"
	"strings"
)

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
		validation.Field(&body.PlayerID, appendRules(rules.PlayerIDRules, validation.Required)...),
		validation.Field(&body.Platform, appendRules(rules.PlatformRules, validation.Required)...),
		validation.Field(&body.Reason, appendRules(rules.InfractionReasonRules, validation.Required)...),
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
		validation.Field(&body.PlayerID, appendRules(rules.PlayerIDRules, validation.Required)...),
		validation.Field(&body.Platform, appendRules(rules.PlatformRules, validation.Required)...),
		validation.Field(&body.Reason, appendRules(rules.InfractionReasonRules, validation.Required)...),
		validation.Field(&body.Duration, appendRules(rules.InfractionDurationRules, validation.Required)...),
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
		validation.Field(&body.PlayerID, appendRules(rules.PlayerIDRules, validation.Required)...),
		validation.Field(&body.Platform, appendRules(rules.PlatformRules, validation.Required)...),
		validation.Field(&body.Reason, appendRules(rules.InfractionReasonRules, validation.Required)...),
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
		validation.Field(&body.PlayerID, appendRules(rules.PlayerIDRules, validation.Required)...),
		validation.Field(&body.Platform, appendRules(rules.PlatformRules, validation.Required)...),
		validation.Field(&body.Reason, appendRules(rules.InfractionReasonRules, validation.Required)...),
		validation.Field(&body.Duration, appendRules(rules.InfractionDurationRules, validation.Required)...),
	)
}

type UpdateWarningParams struct {
	Reason string `json:"reason" form:"reason"`
}

type UpdateMuteParams struct {
	Reason   string `json:"reason" form:"reason"`
	Duration int    `json:"duration" form:"duration"`
}

type UpdateKickParams struct {
	Reason string `json:"reason" form:"reason"`
}

type UpdateBanParams struct {
	Reason   string `json:"reason" form:"reason"`
	Duration int    `json:"duration" form:"duration"`
}

func (body UpdateBanParams) Validate() error {
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		validation.Field(&body.Reason, appendRules(rules.InfractionReasonRules)...),
	)
}
