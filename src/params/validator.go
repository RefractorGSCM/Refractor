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
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"net/http"
	"strings"
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

// stringPointerNotEmpty is a custom validation rule which first checks if the string pointer is nil. If it is, it does
// not return an error since this is expected behaviour. If it's not nil, it checks if it's empty. If it isn't empty, it
// does not return an error. If it is empty, it returns an error.
func stringPointerNotEmpty(val interface{}) error {
	str, _ := val.(*string)

	if str == nil {
		return nil
	}

	if strings.TrimSpace(*str) == "" {
		return fmt.Errorf("cannot be an empty string")
	}

	return nil
}
