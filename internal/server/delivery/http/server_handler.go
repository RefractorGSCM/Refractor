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
	"Refractor/authcheckers"
	"Refractor/domain"
	"Refractor/params"
	"Refractor/pkg/api"
	"Refractor/pkg/api/middleware"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type serverHandler struct {
	service    domain.ServerService
	authorizer domain.Authorizer
	logger     *zap.Logger
}

func ApplyServerHandler(apiGroup *echo.Group, s domain.ServerService, a domain.Authorizer, mware domain.Middleware, log *zap.Logger) {
	handler := &serverHandler{
		service:    s,
		authorizer: a,
		logger:     log,
	}

	// Create the server routing group
	serverGroup := apiGroup.Group("/servers", mware.ProtectMiddleware, mware.ActivationMiddleware)

	// Create an enforcer to authorize the user on the various endpoints
	enforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type: domain.AuthObjRefractor,
	}, log)

	serverGroup.GET("/", handler.GetServers)
	serverGroup.POST("/", handler.CreateServer, enforcer.CheckAuth(authcheckers.RequireAdmin))
}

type resServer struct {
	ID         int64     `json:"id"`
	Game       string    `json:"game"`
	Name       string    `json:"string"`
	Address    string    `json:"address"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

func (h *serverHandler) CreateServer(c echo.Context) error {
	// Validate request body
	var body params.CreateServerParams
	if err := c.Bind(&body); err != nil {
		return err
	}

	if ok, err := api.ValidateRequestBody(body); !ok {
		return err
	}

	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	// Create the new server
	newServer := &domain.Server{
		Game:         body.Game,
		Name:         body.Name,
		Address:      body.Address,
		RCONPort:     body.RCONPort,
		RCONPassword: body.RCONPassword,
	}

	if err := h.service.Store(c.Request().Context(), newServer); err != nil {
		return err
	}

	h.logger.Info("Server created",
		zap.Int64("Server ID", newServer.ID),
		zap.String("Created By", user.Identity.Id),
	)

	return c.JSON(http.StatusCreated, &domain.Response{
		Success: true,
		Message: "Server created",
		Payload: newServer,
	})
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
