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

func (p *Permissions) String() string {
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

func (p *Permissions) Value() *big.Int {
	return p.value
}

// Or runs a bitwise OR comparison between the struct's permissions value and the passed in permissions value.
// The result of this comparison is returned as a new *Permissions instance.
func (p *Permissions) Or(permissions *Permissions) *Permissions {
	baseVal := p.value
	cmpVal := permissions.value

	return newPermission(new(big.Int).Or(baseVal, cmpVal))
}

// ComputeAllowOverrides computes the ALLOW overrides which are stored in a string.
// If the base permissions value has a flag UNSET and there is a present override for it, it will be SET.
// The computed overrides are returned as a new *Permissions instance.
func (p *Permissions) ComputeAllowOverrides(overrideStr string) (*Permissions, error) {
	base := p.value

	overridesP, err := FromString(overrideStr)
	if err != nil {
		return nil, err
	}

	overrides := overridesP.value

	output := new(big.Int).Or(base, overrides)

	return newPermission(output), nil
}

// ComputeDenyOverrides computes the DENY overrides which are stored in a string.
// If the base permissions value has a flag SET and there is a corresponding override for it, it will be UNSET.
// The computed overrides are returned as a new *Permissions instance.
func (p *Permissions) ComputeDenyOverrides(overrideStr string) (*Permissions, error) {
	base := p.value

	overridesP, err := FromString(overrideStr)
	if err != nil {
		return nil, err
	}

	overrides := overridesP.value

	notOverrides := new(big.Int).Not(overrides)
	output := new(big.Int).And(base, notOverrides)

	return newPermission(output), nil
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
