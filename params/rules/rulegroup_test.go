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
	"github.com/franela/goblin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	. "github.com/onsi/gomega"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Rule group", func() {
		g.Describe("Prepend()", func() {
			var group RuleGroup

			g.BeforeEach(func() {
				group = RuleGroup{
					validation.Length(1, 32),
					is.Email,
				}
			})

			g.It("Should add a new rule to the front of the group", func() {
				expected := RuleGroup{
					validation.Required,
					validation.Length(1, 32),
					is.Email,
				}

				output := group.Prepend(validation.Required)

				Expect(output).To(Equal(expected))
			})

			g.It("Should append multiple new rules to the front of the group, matching the order in which they are provided", func() {
				expected := RuleGroup{
					validation.Required,
					validation.NotNil,
					validation.Length(1, 32),
					is.Email,
				}

				output := group.Prepend(validation.Required, validation.NotNil)

				Expect(output).To(Equal(expected))
			})

			g.It("Should not modify the original group", func() {
				original := RuleGroup{
					validation.Length(1, 32),
					is.Email,
				}

				_ = group.Prepend(validation.Required)

				Expect(group).To(Equal(original))
			})
		})

		g.Describe("Append()", func() {
			var group RuleGroup

			g.BeforeEach(func() {
				group = RuleGroup{
					validation.Length(1, 32),
					is.Email,
				}
			})

			g.It("Should add a new rule to the back of the group", func() {
				expected := RuleGroup{
					validation.Length(1, 32),
					is.Email,
					validation.Required,
				}

				output := group.Append(validation.Required)

				Expect(output).To(Equal(expected))
			})

			g.It("Should not modify the original group", func() {
				original := RuleGroup{
					validation.Length(1, 32),
					is.Email,
				}

				_ = group.Append(validation.Required)

				Expect(group).To(Equal(original))
			})
		})
	})
}
