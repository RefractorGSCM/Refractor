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
	"github.com/go-ozzo/ozzo-validation/is"
	"strings"
)

type CreateUserParams struct {
	Username string `json:"username" form:"username"`
	Email    string `json:"email" form:"email"`
}

func (body CreateUserParams) Validate() error {
	body.Username = strings.TrimSpace(body.Username)
	body.Email = strings.TrimSpace(body.Email)

	return ValidateStruct(&body,
		validation.Field(&body.Username, validation.Required, validation.Length(1, 20)),
		validation.Field(&body.Email, validation.Required, is.Email),
	)
}

type LinkPlayerParams struct {
	UserID   string `json:"user_id" form:"user_id"`
	Platform string `json:"platform" form:"platform"`
	PlayerID string `json:"player_id" form:"player_id"`
}

func (body LinkPlayerParams) Validate() error {
	body.UserID = strings.TrimSpace(body.UserID)
	body.Platform = strings.TrimSpace(body.Platform)
	body.PlayerID = strings.TrimSpace(body.PlayerID)

	return ValidateStruct(&body,
		validation.Field(&body.UserID, rules.UserIDRules.Prepend(validation.Required)...),
		validation.Field(&body.Platform, rules.PlatformRules.Prepend(validation.Required)...),
		validation.Field(&body.PlayerID, rules.PlayerIDRules.Prepend(validation.Required)...))
}
