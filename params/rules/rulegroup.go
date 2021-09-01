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

package rules

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

type RuleGroup []validation.Rule

// Prepend adds additional rules to the start of the RuleGroup. It does not modify the original, it simply returns a
// copy with the additional rules prepended.
func (rg RuleGroup) Prepend(rules ...validation.Rule) RuleGroup {
	tmp := make([]validation.Rule, len(rules), len(rules)+len(rg))
	copy(tmp, rules)
	return append(tmp, rg...)
}

// Append adds additional rules to the end of the RuleGroup. It does not modify the original, it simply returns a
// copy with the additional rules appended.
func (rg RuleGroup) Append(rules ...validation.Rule) RuleGroup {
	tmp := make([]validation.Rule, len(rg), len(rg)+len(rules))
	copy(tmp, rg)
	return append(tmp, rules...)
}
