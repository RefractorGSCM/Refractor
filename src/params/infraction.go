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
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"strings"
)

type CreateWarningParams struct {
	PlayerID    string                   `json:"player_id" form:"player_id"`
	Platform    string                   `json:"platform" form:"platform"`
	Reason      string                   `json:"reason" form:"reason"`
	Attachments []CreateAttachmentParams `json:"attachments"`
}

var attachmentArrValidator = validation.Each(validation.By(func(value interface{}) error {
	aBody, ok := value.(CreateAttachmentParams)
	if !ok {
		return fmt.Errorf("could not cast to *domain.Attachment")
	}

	return validation.ValidateStruct(&aBody,
		validation.Field(&aBody.URL, rules.AttachmentURLRules.Prepend(validation.Required)...),
		validation.Field(&aBody.Note, rules.AttachmentNoteRules...))
}))

func (body CreateWarningParams) Validate() error {
	body.PlayerID = strings.TrimSpace(body.PlayerID)
	body.Platform = strings.TrimSpace(body.Platform)
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		validation.Field(&body.PlayerID, rules.PlayerIDRules.Prepend(validation.Required)...),
		validation.Field(&body.Platform, rules.PlatformRules.Prepend(validation.Required)...),
		validation.Field(&body.Reason, rules.InfractionReasonRules.Prepend(validation.Required)...),
		validation.Field(&body.Attachments, attachmentArrValidator),
	)
}

type CreateMuteParams struct {
	PlayerID    string                   `json:"player_id" form:"player_id"`
	Platform    string                   `json:"platform" form:"platform"`
	Reason      string                   `json:"reason" form:"reason"`
	Duration    int                      `json:"duration" form:"duration"`
	Attachments []CreateAttachmentParams `json:"attachments"`
}

func (body CreateMuteParams) Validate() error {
	body.PlayerID = strings.TrimSpace(body.PlayerID)
	body.Platform = strings.TrimSpace(body.Platform)
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		validation.Field(&body.PlayerID, rules.PlayerIDRules.Prepend(validation.Required)...),
		validation.Field(&body.Platform, rules.PlatformRules.Prepend(validation.Required)...),
		validation.Field(&body.Reason, rules.InfractionReasonRules.Prepend(validation.Required)...),
		validation.Field(&body.Duration, rules.InfractionDurationRules.Prepend(validation.Required)...),
		validation.Field(&body.Attachments, attachmentArrValidator),
	)
}

type CreateKickParams struct {
	PlayerID    string                   `json:"player_id" form:"player_id"`
	Platform    string                   `json:"platform" form:"platform"`
	Reason      string                   `json:"reason" form:"reason"`
	Attachments []CreateAttachmentParams `json:"attachments"`
}

func (body CreateKickParams) Validate() error {
	body.PlayerID = strings.TrimSpace(body.PlayerID)
	body.Platform = strings.TrimSpace(body.Platform)
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		validation.Field(&body.PlayerID, rules.PlayerIDRules.Prepend(validation.Required)...),
		validation.Field(&body.Platform, rules.PlatformRules.Prepend(validation.Required)...),
		validation.Field(&body.Reason, rules.InfractionReasonRules.Prepend(validation.Required)...),
		validation.Field(&body.Attachments, attachmentArrValidator),
	)
}

type CreateBanParams struct {
	PlayerID    string                   `json:"player_id" form:"player_id"`
	Platform    string                   `json:"platform" form:"platform"`
	Reason      string                   `json:"reason" form:"reason"`
	Duration    int                      `json:"duration" form:"duration"`
	Attachments []CreateAttachmentParams `json:"attachments"`
}

func (body CreateBanParams) Validate() error {
	body.PlayerID = strings.TrimSpace(body.PlayerID)
	body.Platform = strings.TrimSpace(body.Platform)
	body.Reason = strings.TrimSpace(body.Reason)

	return ValidateStruct(&body,
		validation.Field(&body.PlayerID, rules.PlayerIDRules.Prepend(validation.Required)...),
		validation.Field(&body.Platform, rules.PlatformRules.Prepend(validation.Required)...),
		validation.Field(&body.Reason, rules.InfractionReasonRules.Prepend(validation.Required)...),
		validation.Field(&body.Duration, rules.InfractionDurationRules.Prepend(validation.Required)...),
		validation.Field(&body.Attachments, attachmentArrValidator),
	)
}

type UpdateInfractionParams struct {
	Reason   *string `json:"reason" form:"reason"`
	Duration *int    `json:"duration" form:"duration"`
}

func (body UpdateInfractionParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.Reason, rules.InfractionReasonRules.Prepend(validation.By(stringPointerNotEmpty))...),
		validation.Field(&body.Duration, rules.InfractionDurationRules...),
	)
}
