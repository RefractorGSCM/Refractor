package bitperms

import (
	"fmt"
	"github.com/pkg/errors"
	"math/big"
)

const opTag = "BitPerms."

var errParseString = fmt.Errorf("unable to parse string")

// Permissions is a value which represents a set of flags
type Permissions struct {
	value *big.Int
}

func FromString(str string) (*Permissions, error) {
	const op = opTag + "FromString"

	p := new(big.Int)
	p, ok := p.SetString(str, 10)
	if !ok {
		return nil, errors.Wrap(errParseString, op)
	}

	return newPermission(p), nil
}

func (p *Permissions) ToString() string {
	return p.value.String()
}

// CheckFlag returns true if the permissions value has a flag set.
// This is determined via a bitwise AND operation between the permission value and the flag.
func (p *Permissions) CheckFlag(flag *big.Int) bool {
	return big.NewInt(0).And(p.value, flag).Cmp(flag) == 0
}

// CheckFlags returns true if the permissions value has a flag set.
// This is determined via a bitwise AND operation between the permission value and the flag.
//
// CheckFlags only returns true if ALL of the provided flags are set.
func (p *Permissions) CheckFlags(flags ...*big.Int) bool {
	for _, flag := range flags {
		if big.NewInt(0).And(p.value, flag).Cmp(flag) != 0 {
			return false
		}
	}

	return true
}

// GetFlag returns a big.Int flag which can be used for bitwise comparisons.
// step is the amount of left shifts which should be done to create this flag.
//
// Because the intended use of this package is for each flag to be a value of 1 left shifted n times
// this function serves as a helper so that the verbose big.Int process doesn't need to be repeated
// over and over again. The big.Int process looks like this:
//
// big.NewInt(0).Lsh(big.NewInt(1), 7) which is 1 shifted left 7 times.
func GetFlag(step uint) *big.Int {
	return big.NewInt(0).Lsh(big.NewInt(1), step)
}

func newPermission(n *big.Int) *Permissions {
	return &Permissions{
		value: n,
	}
}

// PermissionBuilder is a builder utility for chaining together flags to build permission values.
type PermissionBuilder struct {
	perm *Permissions
}

func NewPermissionBuilder() *PermissionBuilder {
	return &PermissionBuilder{perm: newPermission(big.NewInt(0))}
}

func (pb *PermissionBuilder) AddFlag(flag *big.Int) *PermissionBuilder {
	newVal := big.NewInt(0).Or(pb.perm.value, flag)

	pb.perm = newPermission(newVal)
	return pb
}

func (pb *PermissionBuilder) GetPermission() *Permissions {
	return pb.perm
}

// Utility functions
func isPowerOfTwo(x *big.Int) bool {
	isNotZero := x.Cmp(big.NewInt(0)) != 0

	// x - 1
	xMinOne := big.NewInt(0).Sub(x, big.NewInt(1))
	xAndMin1 := big.NewInt(0).And(x, xMinOne)

	// if x != 0 && (x & (x - 1)) == 0 then this is a power of two
	return isNotZero && xAndMin1.Cmp(big.NewInt(0)) == 0
}
