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
	"math"
	"regexp"
)

type GroupParams struct {
	Name        string `json:"name" form:"name"`
	Color       int    `json:"color" form:"color"`
	Position    int    `json:"position" form:"position"`
	Permissions string `json:"permissions" form:"permissions"`
}

const maxColor = 0xffffff

var permissionsPattern = regexp.MustCompile("^[0-9]{1,20}$") // numbers only, max length 20

func (body GroupParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.Name, validation.Required, validation.Length(1, 20)),
		validation.Field(&body.Color, validation.Required, validation.Min(0), validation.Max(maxColor)),
		validation.Field(&body.Position, validation.Required, validation.Min(1), validation.Max(math.MaxInt32)),
		validation.Field(&body.Permissions, validation.Required, validation.Match(permissionsPattern)),
	)
}
