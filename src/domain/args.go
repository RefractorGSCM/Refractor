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

package domain

import (
	"fmt"
	"strings"
)

type UpdateArgs map[string]interface{}
type FindArgs map[string]interface{}

func (args UpdateArgs) FilterEmptyStrings() UpdateArgs {
	for key, val := range args {
		fmt.Println(key, val)
		var value string

		// Check if this value is a string or a string pointer. If it isn't, skip it.
		str, ok := val.(string)
		if ok {
			value = str
		} else {
			strPtr, ok := val.(*string)
			if ok {
				value = *strPtr
			}
		}

		// If this value is a string, check if it's empty. If it is, remove it from the map.
		if strings.TrimSpace(value) == "" {
			delete(args, key)
		}
	}

	return args
}
