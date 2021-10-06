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
)

type CreateAttachmentParams struct {
	URL  string `json:"url" form:"url"`
	Note string `json:"note" form:"note"`
}

func (body CreateAttachmentParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.URL, rules.AttachmentURLRules.Prepend(validation.Required)...),
		validation.Field(&body.Note, rules.AttachmentNoteRules...),
	)
}
