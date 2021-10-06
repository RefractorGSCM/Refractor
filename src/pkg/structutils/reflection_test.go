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
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"testing"
)

type testType struct {
	A *int
	B *string
	C *float64
	D *string
}

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	a := 1
	b := "test"
	c := 3.14

	var testStructPtr = &testType{
		A: &a,
		B: &b,
		C: &c,
		D: nil,
	}

	var testStruct = testType{
		A: &a,
		B: &b,
		C: &c,
		D: nil,
	}

	var expectedOutput = map[string]interface{}{
		"A": &a,
		"B": &b,
		"C": &c,
	}

	g.Describe("GetNonNilFieldMap()", func() {
		g.Describe("A valid struct value was passed in", func() {
			g.It("Should not return an error", func() {
				_, err := GetNonNilFieldMap(testStruct)

				Expect(err).To(BeNil())
			})

			g.It("Should return a map containing the non-nil keys and values", func() {
				output, _ := GetNonNilFieldMap(testStructPtr)

				Expect(output).To(Equal(expectedOutput))
			})
		})

		g.Describe("A valid struct pointer was passed in", func() {
			g.It("Should not return an error", func() {
				_, err := GetNonNilFieldMap(testStructPtr)

				Expect(err).To(BeNil())
			})

			g.It("Should return a map containing the non-nil keys and values", func() {
				output, _ := GetNonNilFieldMap(testStructPtr)

				Expect(output).To(Equal(expectedOutput))
			})
		})

		g.Describe("A non struct or struct pointer type was passed in", func() {
			g.It("Should return an error", func() {
				_, err := GetNonNilFieldMap("invalid type")

				Expect(err).ToNot(BeNil())
			})

			g.It("Should return an empty map", func() {
				output, _ := GetNonNilFieldMap("invalid type")

				Expect(output).To(Equal(map[string]interface{}{}))
			})
		})
	})
}
