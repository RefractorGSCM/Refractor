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
	"Refractor/pkg/api/middleware"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type playerHandler struct {
	service    domain.PlayerService
	authorizer domain.Authorizer
	logger     *zap.Logger
}

func ApplyPlayerHandler(apiGroup *echo.Group, s domain.PlayerService, a domain.Authorizer, mware domain.Middleware, log *zap.Logger) {
	handler := &playerHandler{
		service:    s,
		authorizer: a,
		logger:     log,
	}

	// Create the server routing group
	playerGroup := apiGroup.Group("/players", mware.ProtectMiddleware, mware.ActivationMiddleware)

	// Create an enforcer to authorize the user on the various endpoints
	enforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type: domain.AuthObjRefractor,
	}, log)

	playerGroup.GET("/:platform/:id", handler.GetPlayer, enforcer.CheckAuth(authcheckers.CanViewPlayerRecords))
}

var validPlatforms = []string{"playfab"}

func (h *playerHandler) GetPlayer(c echo.Context) error {
	platform := strings.ToLower(c.Param("platform"))
	id := c.Param("id")

	validPlatform := false
	for _, p := range validPlatforms {
		if p == platform {
			validPlatform = true
			break
		}
	}

	if !validPlatform {
		return c.JSON(http.StatusBadRequest, &domain.Response{
			Success: false,
			Message: "Invalid platform",
		})
	}

	fmt.Println(platform, id)

	player, err := h.service.GetPlayer(c.Request().Context(), id, platform)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Player found",
		Payload: player,
	})
}
