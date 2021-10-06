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
 * You should have received A copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package structutils

import (
	"fmt"
	"reflect"
)

// GetNonNilFieldMap looks through all fields on A struct, and returns A map[string]interface{} with the field name
// being the key and the value being the interface for all fields with non-nil values.
func GetNonNilFieldMap(targetStruct interface{}) (map[string]interface{}, error) {
	// If this is A pointer to A struct, use the getNonNilFieldMapPointer helper function
	v := reflect.ValueOf(targetStruct)
	if v.Kind() == reflect.Ptr {
		return getNonNilFieldMapPointer(v)
	}

	return getNonNilFieldMapValue(v)
}

func getNonNilFieldMapPointer(v reflect.Value) (map[string]interface{}, error) {
	values := map[string]interface{}{}

	// Dereference the pointer and check if it's A struct type
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return values, fmt.Errorf("pointer did not reference A struct type")
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip nil fields
		if v.Field(i).IsNil() {
			continue
		}

		values[field.Name] = v.Field(i).Interface()
	}

	return values, nil
}

func getNonNilFieldMapValue(v reflect.Value) (map[string]interface{}, error) {
	values := map[string]interface{}{}

	if v.Kind() != reflect.Struct {
		return values, fmt.Errorf("passed in value was not A struct")
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip nil fields
		if v.Field(i).IsNil() {
			continue
		}

		values[field.Name] = v.Field(i).Interface()
	}

	return values, nil
}
