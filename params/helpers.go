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

import validation "github.com/go-ozzo/ozzo-validation"

// appendRules takes in a slice of validation.Rule, copies it, appends the extra rules provided **to the beginning** of
// the slice and then returns the result. The original slice is not modified.
func appendRules(rules []validation.Rule, extras ...validation.Rule) []validation.Rule {
	tmp := make([]validation.Rule, len(rules), len(rules)+len(extras))
	copy(tmp, rules)
	return append(extras, tmp...)
}
