package bitperms

import (
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"math/big"
	"strconv"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Permissions", func() {
		var i int64
		var iString string

		g.Describe("FromString()", func() {
			g.Describe("A valid string was provided", func() {
				g.Before(func() {
					i = 1 << 32 // use a power of 2 to test since this is typically how bitperms will be used
					iString = strconv.FormatInt(i, 10)
				})

				g.It("Should not return an error", func() {
					_, err := FromString(iString)

					Expect(err).To(BeNil())
				})

				g.It("Permissions should hold the correct integer value", func() {
					perm, _ := FromString(iString)

					iBig := big.NewInt(i)

					Expect(perm.value.Cmp(iBig)).To(Equal(0))
				})
			})

			g.Describe("An invalid string was provided", func() {
				g.Before(func() {
					iString = "notanint_sdjkdhduxwgdjf"
				})

				g.It("Should return an error", func() {
					_, err := FromString(iString)

					Expect(err).ToNot(BeNil())
				})
			})
		})

		g.Describe("ToString()", func() {
			var perm *Permissions

			g.Before(func() {
				i = 1 << 32 // use a power of 2 to test since this is typically how bitperms will be used
				iString = strconv.FormatInt(i, 10)

				perm = &Permissions{
					value: big.NewInt(i),
				}
			})

			g.It("Should return the correct string value", func() {
				Expect(perm.ToString()).To(Equal(iString))
			})
		})

		g.Describe("CheckFlag()", func() {
			var perm *Permissions
			var flag1 *big.Int
			var flag2 *big.Int
			var flag3 *big.Int
			var flag4 *big.Int
			var flag5 *big.Int
			var flag6 *big.Int

			g.Before(func() {
				flag1 = big.NewInt(0).Lsh(big.NewInt(1), 0)
				flag2 = big.NewInt(0).Lsh(big.NewInt(1), 1)
				flag3 = big.NewInt(0).Lsh(big.NewInt(1), 2)
				flag4 = big.NewInt(0).Lsh(big.NewInt(1), 4)
				flag5 = big.NewInt(0).Lsh(big.NewInt(1), 3)
				flag6 = big.NewInt(0).Lsh(big.NewInt(1), 5)

				permVal := big.NewInt(0).Or(flag1, flag2)
				permVal = big.NewInt(0).Or(permVal, flag3)
				permVal = big.NewInt(0).Or(permVal, flag4)

				perm = &Permissions{value: permVal}
			})

			g.Describe("Flag is set", func() {
				g.It("Should return true", func() {
					Expect(perm.CheckFlag(flag1)).To(BeTrue())
					Expect(perm.CheckFlag(flag2)).To(BeTrue())
					Expect(perm.CheckFlag(flag3)).To(BeTrue())
					Expect(perm.CheckFlag(flag4)).To(BeTrue())
				})
			})

			g.Describe("Flag is not set", func() {
				g.It("Should return false", func() {
					Expect(perm.CheckFlag(flag5)).To(BeFalse())
					Expect(perm.CheckFlag(flag6)).To(BeFalse())
				})
			})
		})

		g.Describe("CheckFlags()", func() {
			var perm *Permissions
			var flag1 *big.Int
			var flag2 *big.Int
			var flag3 *big.Int
			var flag4 *big.Int
			var flag5 *big.Int
			var flag6 *big.Int

			g.Before(func() {
				flag1 = big.NewInt(0).Lsh(big.NewInt(1), 0)
				flag2 = big.NewInt(0).Lsh(big.NewInt(1), 1)
				flag3 = big.NewInt(0).Lsh(big.NewInt(1), 2)
				flag4 = big.NewInt(0).Lsh(big.NewInt(1), 4)
				flag5 = big.NewInt(0).Lsh(big.NewInt(1), 3)
				flag6 = big.NewInt(0).Lsh(big.NewInt(1), 5)

				permVal := big.NewInt(0).Or(flag1, flag2)
				permVal = big.NewInt(0).Or(permVal, flag3)
				permVal = big.NewInt(0).Or(permVal, flag4)

				perm = &Permissions{value: permVal}
			})

			g.Describe("Flags are set", func() {
				g.It("Should return true", func() {
					Expect(perm.CheckFlags(flag1, flag2, flag3, flag4))
				})
			})

			g.Describe("Not all flags are set", func() {
				g.It("Should return false", func() {
					Expect(perm.CheckFlags(flag1, flag2, flag3, flag5, flag6)).To(BeFalse())
				})
			})

			g.Describe("No flags are set", func() {
				g.It("Should return false", func() {
					Expect(perm.CheckFlags(flag5, flag6)).To(BeFalse())
				})
			})
		})

		g.Describe("isPowerOfTwo()", func() {
			g.It("Should return true", func() {
				x := big.NewInt(0).Lsh(big.NewInt(1), 32) // 1 << 32
				Expect(isPowerOfTwo(x)).To(BeTrue())

				x = big.NewInt(0).Lsh(big.NewInt(1), 64) // 1 << 64
				Expect(isPowerOfTwo(x)).To(BeTrue())

				x = big.NewInt(0).Lsh(big.NewInt(1), 128) // 1 << 128
				Expect(isPowerOfTwo(x)).To(BeTrue())

				x = big.NewInt(0).Lsh(big.NewInt(1), 256) // 1 << 256
				Expect(isPowerOfTwo(x)).To(BeTrue())

				x = big.NewInt(0).Lsh(big.NewInt(1), 512) // 1 << 512
				Expect(isPowerOfTwo(x)).To(BeTrue())

				x = big.NewInt(0).Lsh(big.NewInt(1), 3) // 1 << 3
				Expect(isPowerOfTwo(x)).To(BeTrue())

				x = big.NewInt(0).Lsh(big.NewInt(1), 5) // 1 << 5
				Expect(isPowerOfTwo(x)).To(BeTrue())

				x = big.NewInt(0).Lsh(big.NewInt(1), 7) // 1 << 7
				Expect(isPowerOfTwo(x)).To(BeTrue())
			})

			g.It("Should return false", func() {
				Expect(isPowerOfTwo(big.NewInt(0))).To(BeFalse())
				Expect(isPowerOfTwo(big.NewInt(3))).To(BeFalse())
				Expect(isPowerOfTwo(big.NewInt(6))).To(BeFalse())
				Expect(isPowerOfTwo(big.NewInt(9))).To(BeFalse())
				Expect(isPowerOfTwo(big.NewInt(5423543255))).To(BeFalse())
				Expect(isPowerOfTwo(big.NewInt(3646278))).To(BeFalse())
				Expect(isPowerOfTwo(big.NewInt(1798223))).To(BeFalse())
			})
		})

		g.Describe("GetFlag()", func() {
			g.It("Returns the correct flag for each provided step", func() {

				var step uint = 1
				for ; step < 512; step++ {
					expected := big.NewInt(0).Lsh(big.NewInt(1), step)
					Expect(GetFlag(step).Cmp(expected) == 0).To(BeTrue())
				}
			})
		})
	})

	g.Describe("PermissionBuilder", func() {
		var pb *PermissionBuilder

		g.BeforeEach(func() {
			pb = NewPermissionBuilder()
		})

		g.Describe("NewPermissionBuilder()", func() {
			g.It("Should return a new permission builder with a permission value of 0", func() {
				Expect(pb.perm.value.Cmp(big.NewInt(0)) == 0).To(BeTrue())
			})
		})

		g.Describe("AddFlag()", func() {
			g.It("Should add a new flag", func() {
				flag := big.NewInt(0).Lsh(big.NewInt(1), 1) // 1 << 1
				pb.AddFlag(flag)

				newPermVal := pb.perm.value
				Expect(big.NewInt(0).And(newPermVal, flag).Cmp(flag) == 0).To(BeTrue()) // (newPermVal & flag) == 0
			})

			g.It("Should allow chaining of flags", func() {
				flag1 := big.NewInt(0).Lsh(big.NewInt(1), 1) // 1 << 1
				flag2 := big.NewInt(0).Lsh(big.NewInt(1), 2) // 1 << 2
				flag3 := big.NewInt(0).Lsh(big.NewInt(1), 5) // 1 << 5

				pb.AddFlag(flag1).AddFlag(flag2).AddFlag(flag3)

				newPermVal := pb.perm.value
				Expect(big.NewInt(0).And(newPermVal, flag1).Cmp(flag1) == 0).To(BeTrue()) // (newPermVal & flag1) == 0
				Expect(big.NewInt(0).And(newPermVal, flag2).Cmp(flag2) == 0).To(BeTrue()) // (newPermVal & flag2) == 0
				Expect(big.NewInt(0).And(newPermVal, flag3).Cmp(flag3) == 0).To(BeTrue()) // (newPermVal & flag3) == 0
			})
		})

		g.Describe("GetPermission()", func() {
			g.It("Should return the correct final permission value", func() {
				flag1 := big.NewInt(0).Lsh(big.NewInt(1), 1) // 1 << 1
				pb.AddFlag(flag1)

				flag2 := big.NewInt(0).Lsh(big.NewInt(1), 2) // 1 << 2
				pb.AddFlag(flag2)

				flag3 := big.NewInt(0).Lsh(big.NewInt(1), 5) // 1 << 5
				pb.AddFlag(flag3)

				// Calculate the correct permission value
				expected := big.NewInt(0).Or(flag1, flag2)
				expected = big.NewInt(0).Or(expected, flag3)

				Expect(pb.GetPermission().value.Cmp(expected) == 0).To(BeTrue())
			})
		})

		g.Describe("ComputeAllowOverrides()", func() {
			var base *Permissions
			var overridePerm *Permissions
			var overrideStr string

			g.BeforeEach(func() {
				// Create base permission value
				flag1 := big.NewInt(0).Lsh(big.NewInt(1), 0)  // 1 << 0
				flag2 := big.NewInt(0).Lsh(big.NewInt(1), 1)  // 1 << 1
				flag3 := big.NewInt(0).Lsh(big.NewInt(1), 4)  // 1 << 4
				flag4 := big.NewInt(0).Lsh(big.NewInt(1), 12) // 1 << 12

				basePermVal := new(big.Int).Or(flag1, flag2)
				basePermVal = basePermVal.Or(basePermVal, flag3)
				basePermVal = basePermVal.Or(basePermVal, flag4)

				base = newPermission(basePermVal)

				// Create override permission string value
				oFlag1 := big.NewInt(0).Lsh(big.NewInt(1), 2) // 1 << 2
				oFlag2 := big.NewInt(0).Lsh(big.NewInt(1), 3) // 1 << 3
				oFlag3 := big.NewInt(0).Lsh(big.NewInt(1), 9) // 1 << 9

				overrideVal := new(big.Int).Or(oFlag1, oFlag2)
				overrideVal = overrideVal.Or(overrideVal, oFlag3)
				overridePerm = newPermission(overrideVal)
				overrideStr = overrideVal.String()
			})

			g.Describe("Valid overrides string provided", func() {
				g.It("Should not return an error", func() {
					_, err := base.ComputeAllowOverrides(overrideStr)

					Expect(err).To(BeNil())
				})

				g.It("Should return the correct computed permissions value", func() {
					computed, err := base.ComputeAllowOverrides(overrideStr)

					expected := new(big.Int).Or(base.value, overridePerm.value)

					Expect(err).To(BeNil())
					Expect(computed.value.Cmp(expected) == 0).To(BeTrue())
				})
			})

			g.Describe("An invalid overrides string was provided", func() {
				g.It("Should return an error", func() {
					_, err := base.ComputeAllowOverrides("invalidstr")

					Expect(err).ToNot(BeNil())
				})
			})
		})
	})
}
