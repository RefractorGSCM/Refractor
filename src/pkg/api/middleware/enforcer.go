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

package middleware

import (
	"Refractor/domain"
	"Refractor/pkg/api"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type Enforcer struct {
	authorizer domain.Authorizer
	scope      domain.AuthScope
	logger     *zap.Logger
}

func NewEnforcer(authorizer domain.Authorizer, scope domain.AuthScope, log *zap.Logger) *Enforcer {
	return &Enforcer{
		authorizer: authorizer,
		scope:      scope,
		logger:     log,
	}
}

// CheckAuth takes in an authorization checker (domain.AuthChecker) and uses it and other factors to determine if the
// user is authorized to access an endpoint. The following checks are performed in the following order:
//
// 1. Grant access if the user is a super admin.
//
// 2. Grant access if the authorization checker returns true.
//
// CheckAuth must be placed after the protection middleware (see middleware.NewAPIProtectMiddleware) because it relies
// on the "user" being set in context. CheckAuth leverages the authorizer to perform the actual authorization tests
// depending on scope.
//
// If the scope is set to domain.AuthObjServer and no scope ID was provided, it will automatically attempt to parse
// the server's ID from the :id url parameter and attach it to the auth scope.
func (e *Enforcer) CheckAuth(authChecker domain.AuthChecker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the user
			user, ok := c.Get("user").(*domain.AuthUser)
			if !ok {
				return c.JSON(http.StatusUnauthorized, domain.ResponseUnauthorized)
			}

			ctx := c.Request().Context()

			switch e.scope.Type {
			case domain.AuthObjServer:
				var idField = "id"

				// Check if a server ID override was set
				if e.scope.IDFieldName != "" {
					idField = e.scope.IDFieldName
				}

				serverID, ok := parseID(c.Param(idField))

				if !ok {
					return c.JSON(http.StatusBadRequest, &domain.Response{
						Success: false,
						Message: "An invalid server ID provided",
					})
				}

				e.scope.ID = serverID
			}

			// Use the authorizer and the api.CheckPermissions function to check if the user has permission
			// to access this endpoint. We use api.CheckPermissions because it automatically checks if the user is a
			// super admin and returns true if they do.
			hasPermission, err := api.CheckPermissions(ctx, e.authorizer, e.scope, user.Identity.Id, authChecker)
			if err != nil {
				return err
			}

			if !hasPermission {
				return c.JSON(http.StatusUnauthorized, domain.ResponseUnauthorized)
			}

			return next(c)
		}
	}
}

func parseID(idString string) (int64, bool) {
	if idString == "" {
		return 0, false
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}
