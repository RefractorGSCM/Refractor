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

package validators

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
)

func ValueInStrArray(arr []string) validation.RuleFunc {
	return func(value interface{}) error {
		value, _ = value.(string)

		if value == "" {
			return nil
		}

		for _, val := range arr {
			if val == value {
				return nil
			}
		}

		return errors.New(fmt.Sprintf("must be one of: %v", arr))
	}
}

func PtrValueInStrArray(arr []string) validation.RuleFunc {
	return func(value interface{}) error {
		strValPtr, _ := value.(*string)

		if strValPtr == nil {
			return nil
		}

		strVal := *strValPtr

		if strVal == "" {
			return nil
		}

		for _, val := range arr {
			if val == strVal {
				return nil
			}
		}

		return errors.New(fmt.Sprintf("must be one of: %v", arr))
	}
}
