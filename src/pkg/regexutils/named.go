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

package regexutils

import "regexp"

func MapNamedMatches(pattern *regexp.Regexp, data string) map[string]string {
	matches := pattern.FindStringSubmatch(data)

	if len(matches) < 1 {
		return nil
	}

	matches = matches[1:] // skip first match since it's the entire match, not just the submatches

	namedMatches := map[string]string{}

	for i, name := range pattern.SubexpNames() {
		// skip the first global match
		if i == 0 {
			continue
		}

		namedMatches[name] = matches[i-1]
	}

	return namedMatches
}
