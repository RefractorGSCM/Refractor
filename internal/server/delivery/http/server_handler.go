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

package http

import (
	"Refractor/domain"
	"github.com/labstack/echo/v4"
)

type serverHandler struct {
	service domain.ServerService
}

func ApplyServerHandler(apiGroup *echo.Group, s domain.ServerService, authorizer domain.Authorizer, protect echo.MiddlewareFunc) {
	handler := &serverHandler{
		service: s,
	}

	// Create the server routing group
	serverGroup := apiGroup.Group("/servers")

	serverGroup.GET("/", handler.GetServers, protect)
}

// GetServers is the route handler for /api/v1/servers
// It returns a JSON array containing all servers which the requesting user has access to.
func (h *serverHandler) GetServers(c echo.Context) error {
	return c.String(200, "ok")
}
