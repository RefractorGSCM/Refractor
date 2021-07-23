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

package env

import (
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("RequireEnv()", func() {
		g.It("Adds a new missing env variable", func() {
			env := RequireEnv("TEST_1")

			Expect(env.missingVars).To(ContainElement("TEST_1"))
		})

		g.It("Adds a new missing env variable when chained", func() {
			env := RequireEnv("TEST_1").
				RequireEnv("TEST_2")

			Expect(env.missingVars).To(ContainElements("TEST_1", "TEST_2"))
		})

		g.It("Does not add an existing env variable", func() {
			err := os.Setenv("ENV_TEST_VAR_18263782", "true")
			Expect(err).To(BeNil())

			env := RequireEnv("ENV_TEST_VAR_18263782")
			Expect(env.missingVars).ToNot(ContainElement("ENV_TEST_VAR_18263782"))
		})
	})

	g.Describe("GetError()", func() {
		g.It("Outputs an error message when one or more env variables are missing", func() {
			err := RequireEnv("TEST_1").RequireEnv("TEST_2").GetError()

			Expect(err).ToNot(BeNil())
		})

		g.It("Outputs an error which contains a list of missing env variables", func() {
			err := os.Setenv("ENV_TEST_VAR_18263783", "true")
			Expect(err).To(BeNil())

			err = RequireEnv("TEST_1").
				RequireEnv("ENV_TEST_VAR_18263783").
				RequireEnv("TEST_2").
				GetError()

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("TEST_1"))
			Expect(err.Error()).To(ContainSubstring("TEST_2"))
			Expect(err.Error()).ToNot(ContainSubstring("ENV_TEST_VAR_18263783"))
		})
	})
}

