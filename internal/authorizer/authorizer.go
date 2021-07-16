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
	"fmt"
	"math/big"
)

type authorizer struct {
}

func NewAuthorizer() domain.Authorizer {
	return &authorizer{}
}

func (a *authorizer) HasPermission(scope domain.AuthScope, userID string, requiredFlags []*big.Int) (bool, error) {
	switch scope.Type {
	case domain.AuthObjRefractor:
		return a.hasPermissionRefractor(userID, requiredFlags)

	case domain.AuthObjServer:
		serverID, ok := scope.ID.(int64)
		if !ok {
			return false, fmt.Errorf("an invalid server id was provided")
		}

		return a.hasPermissionServer(userID, serverID, requiredFlags)
	}

	return false, fmt.Errorf("an invalid AuthScope.type was provided")
}
