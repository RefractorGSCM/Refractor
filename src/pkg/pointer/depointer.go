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

package pointer

import "reflect"

// DePointer takes in an interface{} and checks if it's a pointer or not. If it is a pointer, it dereferences it and
// returns the dereferenced interface. If it is not a pointer, it just returns the value.
func DePointer(val interface{}) interface{} {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		return v.Elem().Interface()
	}

	return val
}
