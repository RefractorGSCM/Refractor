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
)

type CreateUserParams struct {
	Username string `json:"username" form:"username"`
	Email    string `json:"email" form:"email"`
}

func (body CreateUserParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.Username, validation.Required, validation.Length(1, 20)),
		validation.Field(&body.Email, validation.Required, is.Email),
	)
}
