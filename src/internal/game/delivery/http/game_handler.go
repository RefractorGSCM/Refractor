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
)

type gameHandler struct {
	service domain.GameService
}

func ApplyGameHandler(apiGroup *echo.Group, s domain.GameService, mware domain.Middleware) {
	handler := &gameHandler{
		service: s,
	}

	// Create the server routing group
	gameGroup := apiGroup.Group("/games", mware.ProtectMiddleware, mware.ActivationMiddleware)

	gameGroup.GET("/", handler.GetGames)
}

type resGame struct {
	Name        string `json:"name"`
	Platform    string `json:"platform"`
	ChatEnabled bool   `json:"chat_enabled"`
}

func (h *gameHandler) GetGames(c echo.Context) error {
	allGames := h.service.GetAllGames()

	var res []*resGame
	for _, g := range allGames {
		res = append(res, &resGame{
			Name:        g.GetName(),
			Platform:    g.GetPlatform().GetName(),
			ChatEnabled: g.GetConfig().EnableChat,
		})
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: fmt.Sprintf("Found %d games", len(allGames)),
		Payload: res,
	})
}
