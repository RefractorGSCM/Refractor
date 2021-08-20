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
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"math"
	"net/http"
	"regexp"
)

type CreateGroupParams struct {
	Name        string `json:"name" form:"name"`
	Color       int    `json:"color" form:"color"`
	Position    int    `json:"position" form:"position"`
	Permissions string `json:"permissions" form:"permissions"`
}

const maxColor = 0xffffff

var permissionsPattern = regexp.MustCompile("^[0-9]{1,20}$") // numbers only, max length 20

func (body CreateGroupParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.Name, validation.Required, validation.Length(1, 20)),
		validation.Field(&body.Color, validation.Required, validation.Min(0), validation.Max(maxColor)),
		validation.Field(&body.Position, validation.Required, validation.Min(1), validation.Max(math.MaxInt32)),
		validation.Field(&body.Permissions, validation.Required, validation.Match(permissionsPattern)),
	)
}

type UpdateGroupParams struct {
	Name        *string `json:"name" form:"name"`
	Color       *int    `json:"color" form:"color"`
	Position    *int    `json:"position" form:"position"`
	Permissions *string `json:"permissions" form:"permissions"`
}

func (body UpdateGroupParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.Name, validation.By(stringPointerNotEmpty), validation.Length(1, 20)),
		validation.Field(&body.Color, validation.Min(0), validation.Max(maxColor)),
		validation.Field(&body.Position, validation.Min(1), validation.Max(math.MaxInt32)),
		validation.Field(&body.Permissions, validation.By(stringPointerNotEmpty), validation.Match(permissionsPattern)),
	)
}

type GroupReorderParams struct {
	GroupID int64 `json:"id"`
	NewPos  int64 `json:"pos"`
}

func (body GroupReorderParams) Validate() error {
	return validation.ValidateStruct(&body,
		validation.Field(&body.GroupID, validation.Required, validation.Min(1), validation.Max(math.MaxInt32)),
		validation.Field(&body.NewPos, validation.Required, validation.Min(1), validation.Max(math.MaxInt32-1)), // -1 to account for the base group
	)
}

type GroupReorderArray []GroupReorderParams

type GroupReorderErr struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

func (body GroupReorderArray) Validate() error {
	if len(body) < 1 {
		return wrapError(domain.NewHTTPError(fmt.Errorf("no reorder instructions provided"),
			http.StatusBadRequest, "no reorder instructions provided"))
	}

	// Override default error behaviour since this is an unusual usecase where we need to provide both the index of the
	// element which the error occurred on as well as the normal error message.
	for idx, grp := range body {
		if err := grp.Validate(); err != nil {
			return &domain.HTTPError{
				Cause:   err,
				Message: "Input errors exist",
				ValidationErrors: map[string]string{
					"index":   fmt.Sprintf("%d", idx),
					"message": err.Error(),
				},
				Status: http.StatusBadRequest,
			}
		}
	}

	return nil
}

type SetUserGroupParams struct {
	UserID  string `json:"user_id" form:"user_id"`
	GroupID int64  `json:"group_id" form:"group_id"`
}

func (body SetUserGroupParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.UserID, validation.Required, validation.Length(36, 36)), // uuid length
		validation.Field(&body.GroupID, validation.Required, validation.Min(1), validation.Max(math.MaxInt32)),
	)
}

type SetServerOverrideParams struct {
	DenyOverrides  string `json:"deny_overrides" form:"deny_overrides"`
	AllowOverrides string `json:"allow_overrides" form:"allow_overrides"`
}

func (body SetServerOverrideParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.DenyOverrides, validation.Required, validation.Match(permissionsPattern)),
		validation.Field(&body.AllowOverrides, validation.Required, validation.Match(permissionsPattern)),
	)
}
