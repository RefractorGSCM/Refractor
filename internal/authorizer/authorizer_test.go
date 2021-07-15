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
