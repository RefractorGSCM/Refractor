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

import (
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Whitelist test", func() {
		g.Describe("StringKeyMap", func() {
			g.Describe("FilterKeys()", func() {
				var input map[string]interface{}

				g.BeforeEach(func() {
					input = map[string]interface{}{
						"Key1": "val1",
						"Key2": "val1",
						"Key3": "val1",
						"Key4": "val1",
					}
				})

				g.It("Should remove fields with non-whitelisted keys", func() {
					wl := StringKeyMap{
						"Key1",
						"Key3",
					}

					expected := map[string]interface{}{
						"Key1": "val1",
						"Key3": "val1",
					}

					output := wl.FilterKeys(input)

					Expect(output).To(Equal(expected))
				})
			})
		})
	})
}
