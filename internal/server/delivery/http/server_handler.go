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
	"Refractor/pkg/structutils"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
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
	rEnforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type: domain.AuthObjRefractor,
	}, log)

	sEnforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type: domain.AuthObjServer,
	}, log)

	serverGroup.GET("/", handler.GetServers)
	serverGroup.GET("/:id", handler.GetServerByID, sEnforcer.CheckAuth(authcheckers.CanViewServer))
	serverGroup.POST("/", handler.CreateServer, rEnforcer.CheckAuth(authcheckers.RequireAdmin))
	serverGroup.PATCH("/deactivate/:id", handler.DeactivateServer, rEnforcer.CheckAuth(authcheckers.RequireAdmin))
	serverGroup.PATCH("/:id", handler.UpdateServer, rEnforcer.CheckAuth(authcheckers.RequireAdmin))
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

type resServer struct {
	ID            int64            `json:"id"`
	Game          string           `json:"game"`
	Name          string           `json:"name"`
	Address       string           `json:"address"`
	RCONPort      string           `json:"rcon_port"`
	Deactivated   bool             `json:"deactivated"`
	CreatedAt     time.Time        `json:"created_at"`
	ModifiedAt    time.Time        `json:"modified_at"`
	OnlinePlayers []*domain.Player `json:"online_players"`
	Status        string           `json:"status"`
}

// GetServers is the route handler for /api/v1/servers
// It returns a JSON array containing all servers which the requesting user has access to.
func (h *serverHandler) GetServers(c echo.Context) error {
	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	ctx := context.WithValue(c.Request().Context(), "user", user)
	servers, err := h.service.GetAllAccessible(ctx)
	if err != nil {
		return err
	}

	var resServers []*resServer

	// Transform servers into resServers
	for _, server := range servers {
		resServer := &resServer{
			ID:          server.ID,
			Game:        server.Game,
			Name:        server.Name,
			Address:     server.Address,
			RCONPort:    server.RCONPort,
			Deactivated: server.Deactivated,
			CreatedAt:   server.CreatedAt,
			ModifiedAt:  server.ModifiedAt,
		}

		// Get server's data
		data, err := h.service.GetServerData(server.ID)
		if err == nil {
			for _, op := range data.OnlinePlayers {
				resServer.OnlinePlayers = append(resServer.OnlinePlayers, op)
			}

			if len(resServer.OnlinePlayers) < 1 {
				resServer.OnlinePlayers = []*domain.Player{}
			}

			resServer.Status = data.Status
		} else if errors.Cause(err) != domain.ErrNotFound {
			return err
		}

		resServers = append(resServers, resServer)
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: fmt.Sprintf("Fetched %d servers", len(resServers)),
		Payload: resServers,
	})
}

func (h *serverHandler) GetServerByID(c echo.Context) error {
	serverIDString := c.Param("id")

	serverID, err := strconv.ParseInt(serverIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid server id"), http.StatusBadRequest, "")
	}

	server, err := h.service.GetByID(c.Request().Context(), serverID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Server fetched",
		Payload: &resServer{
			ID:          server.ID,
			Game:        server.Game,
			Name:        server.Name,
			Address:     server.Address,
			RCONPort:    server.RCONPort,
			Deactivated: server.Deactivated,
			CreatedAt:   server.CreatedAt,
			ModifiedAt:  server.ModifiedAt,
		},
	})
}

func (h *serverHandler) DeactivateServer(c echo.Context) error {
	serverIDString := c.Param("id")

	serverID, err := strconv.ParseInt(serverIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid server id"), http.StatusBadRequest, "")
	}

	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	// Deactivate the server
	if err := h.service.Deactivate(c.Request().Context(), serverID); err != nil {
		return err
	}

	h.logger.Info(
		"A server has been deactivated",
		zap.Int64("Server ID", serverID),
		zap.String("Deactivated By", user.Identity.Id),
	)

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Server deactivated",
	})
}

func (h *serverHandler) UpdateServer(c echo.Context) error {
	serverIDString := c.Param("id")

	serverID, err := strconv.ParseInt(serverIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid server id"), http.StatusBadRequest, "")
	}

	// Validate request body
	var body params.UpdateServerParams
	if err := c.Bind(&body); err != nil {
		return err
	}

	if ok, err := api.ValidateRequestBody(body); !ok {
		return err
	}

	// Get update args
	updateArgs, err := structutils.GetNonNilFieldMap(body)
	if err != nil {
		return err
	}

	if len(updateArgs) < 1 {
		return c.JSON(http.StatusBadRequest, &domain.Response{
			Success: false,
			Message: "No update fields provided",
		})
	}

	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	// Update the server
	updated, err := h.service.Update(c.Request().Context(), serverID, updateArgs)
	if err != nil {
		return err
	}

	h.logger.Info(
		"A server has been updated",
		zap.Int64("Server ID", serverID),
		zap.Any("Update Args", updateArgs),
		zap.String("Updated By", user.Identity.Id),
	)

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Server updated",
		Payload: updated,
	})
}
