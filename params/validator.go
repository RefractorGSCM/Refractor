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
	"encoding/json"
	validation "github.com/go-ozzo/ozzo-validation"
	"net/http"
)

// ValidateStruct is a wrapper function which calls the wrapError function around validation.ValidateStruct.
func ValidateStruct(structPtr interface{}, fields ...*validation.FieldRules) error {
	return wrapError(validation.ValidateStruct(structPtr, fields...))
}

func wrapError(err error) error {
	if err == nil {
		return err
	}

	// If err is already an http error, just return it
	if _, ok := err.(*domain.HTTPError); ok {
		return err
	}

	httpError := &domain.HTTPError{
		Cause:   domain.ErrInvalid,
		Message: "Input errors exist",
		Status:  http.StatusBadRequest,
	}

	data, err := json.Marshal(err)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &httpError.ValidationErrors)
	if err != nil {
		return err
	}

	return httpError
}
