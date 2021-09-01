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

package whitelist

type StringMap []string

func (m StringMap) FilterKeys(input map[string]string) map[string]string {
	for k, _ := range input {
		allowedKey := false
		for _, whitelistKey := range m {
			if k == whitelistKey {
				allowedKey = true
				break
			}
		}

		if !allowedKey {
			delete(input, k)
		}
	}

	return input
}
