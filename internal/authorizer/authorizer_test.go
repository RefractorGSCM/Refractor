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

package authorizer

import (
	"Refractor/domain"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"math/big"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("HasPermission()", func() {
		var a *authorizer

		g.BeforeEach(func() {
			a = &authorizer{}
		})

		g.Describe("A valid AuthScope was provided", func() {
			g.It("Should not return an error", func() {
				var as domain.AuthScope
				var err error

				// Refractor auth scope
				as = domain.AuthScope{
					Type: domain.AuthObjRefractor,
					ID:   nil,
				}
				_, err = a.HasPermission(as, "userID", []*big.Int{})
				Expect(err).To(BeNil())

				// Server auth scope
				as = domain.AuthScope{
					Type: domain.AuthObjServer,
					ID:   int64(1),
				}
				_, err = a.HasPermission(as, "userID", []*big.Int{})
				Expect(err).To(BeNil())
			})
		})

		g.Describe("An invalid AuthScope was provided", func() {
			g.It("Should return an error", func() {
				as := domain.AuthScope{
					Type: "Invalid",
				}

				_, err := a.HasPermission(as, "userID", []*big.Int{})

				Expect(err).ToNot(BeNil())
			})
		})
	})
}
