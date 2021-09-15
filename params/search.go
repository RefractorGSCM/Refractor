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
	"Refractor/domain"
	"Refractor/params/rules"
	"Refractor/params/validators"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"strings"
)

type SearchParams struct {
	Offset int `json:"offset" form:"offset"`
	Limit  int `json:"limit" form:"offset"`
}

func (body SearchParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.Offset, rules.SearchOffsetRules.Prepend(validation.Required)...),
		validation.Field(&body.Limit, rules.SearchLimitRules.Prepend(validation.Required)...),
	)
}

type SearchPlayerParams struct {
	Term     string `json:"term" form:"term"`
	Type     string `json:"type" form:"type"`
	Platform string `json:"platform" form:"platform"`
	SearchParams
}

var validPlayerSearchTypes = []string{"name", "id"}

func (body SearchPlayerParams) Validate() error {
	if err := body.SearchParams.Validate(); err != nil {
		return err
	}

	body.Term = strings.TrimSpace(body.Term)
	body.Type = strings.TrimSpace(body.Type)

	return ValidateStruct(&body,
		validation.Field(&body.Term, validation.Required, validation.Length(1, 128)),
		validation.Field(&body.Type, validation.Required, validation.By(validators.ValueInStrArray(validPlayerSearchTypes))),
		validation.Field(&body.Platform, validation.By(validators.ValueInStrArray(domain.AllPlatforms)),
			validation.By(func(value interface{}) error {
				// if body.Type is set to "id", then platform is required
				if body.Type != "id" {
					return nil
				}

				platform, ok := value.(string)
				if !ok || (ok && len(strings.TrimSpace(platform)) == 0) {
					return errors.New("platform is required if search type is id")
				}

				return nil
			})),
	)
}
