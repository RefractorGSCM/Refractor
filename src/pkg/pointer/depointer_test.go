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

import (
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("DePointer()", func() {
		g.Describe("A pointer was passed in", func() {
			g.It("Should return the de-referenced value", func() {
				var testStr = "test"

				res := DePointer(&testStr)

				Expect(res).To(Equal(testStr))
			})
		})

		g.Describe("A non-pointer was passed in", func() {
			g.It("Should return the passed in value", func() {
				var testStr = "test"

				res := DePointer(testStr)

				Expect(res).To(Equal(testStr))
			})
		})
	})
}
