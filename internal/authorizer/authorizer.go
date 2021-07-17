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
	"context"
	"fmt"
)

const opTag = "Authorizer."

type authorizer struct {
	groupRepo domain.GroupRepo
}

func NewAuthorizer(gr domain.GroupRepo) domain.Authorizer {
	return &authorizer{
		groupRepo: gr,
	}
}

func (a *authorizer) HasPermission(ctx context.Context, scope domain.AuthScope, userID string, comparator domain.AuthChecker) (bool, error) {
	switch scope.Type {
	case domain.AuthObjRefractor:
		return a.hasPermissionRefractor(ctx, userID, comparator)

	case domain.AuthObjServer:
		serverID, ok := scope.ID.(int64)
		if !ok {
			return false, fmt.Errorf("an invalid server id was provided")
		}

		return a.hasPermissionServer(ctx, userID, serverID, comparator)
	}

	return false, fmt.Errorf("an invalid AuthScope.type was provided")
}
