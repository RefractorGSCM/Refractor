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
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type serverHandler struct {
	service domain.ServerService
}

func ApplyServerHandler(apiGroup *echo.Group, s domain.ServerService, authorizer domain.Authorizer, mware domain.Middleware) {
	handler := &serverHandler{
		service: s,
	}

	// Create the server routing group
	serverGroup := apiGroup.Group("/servers", mware.ProtectMiddleware, mware.ActivationMiddleware)

	serverGroup.GET("/", handler.GetServers)
}

type resServer struct {
	ID         int64     `json:"id"`
	Game       string    `json:"game"`
	Name       string    `json:"string"`
	Address    string    `json:"address"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

// GetServers is the route handler for /api/v1/servers
// It returns a JSON array containing all servers which the requesting user has access to.
func (h *serverHandler) GetServers(c echo.Context) error {
	servers, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return err
	}

	var resServers []*resServer

	// Transform servers into resServers
	for _, server := range servers {
		resServers = append(resServers, &resServer{
			ID:         server.ID,
			Game:       server.Game,
			Name:       server.Name,
			Address:    server.Address,
			CreatedAt:  server.CreatedAt,
			ModifiedAt: server.ModifiedAt,
		})
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: fmt.Sprintf("Fetched %d servers", len(resServers)),
		Payload: resServers,
	})
}
