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
	"strings"
)

type gameHandler struct {
	service domain.GameService
}

func ApplyGameHandler(apiGroup *echo.Group, s domain.GameService, mware domain.Middleware, a domain.Authorizer, log *zap.Logger) {
	handler := &gameHandler{
		service: s,
	}

	// Create an app enforcer to authorize the user on the various chat endpoints
	enforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type: domain.AuthObjRefractor,
	}, log)

	// Create the server routing group
	gameGroup := apiGroup.Group("/games", mware.ProtectMiddleware, mware.ActivationMiddleware)

	gameGroup.GET("/", handler.GetGames)
	gameGroup.GET("/settings/:game", handler.GetGameSettings, enforcer.CheckAuth(authcheckers.DenyAll))                // super admin only
	gameGroup.PATCH("/settings/:game", handler.SetGameSettings, enforcer.CheckAuth(authcheckers.DenyAll))              // super admin only
	gameGroup.GET("/settings/:game/default", handler.GetDefaultGameSettings, enforcer.CheckAuth(authcheckers.DenyAll)) // super admin only
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

func (h *gameHandler) GetGameSettings(c echo.Context) error {
	gameName := c.Param("game")

	if len(strings.TrimSpace(gameName)) == 0 || !h.service.GameExists(gameName) {
		return c.JSON(http.StatusBadRequest, &domain.Response{
			Success: false,
			Message: "Invalid game",
		})
	}

	// Get the game
	game, err := h.service.GetGame(gameName)
	if err != nil {
		return err
	}

	settings, err := h.service.GetGameSettings(game)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Payload: settings,
	})
}

func (h *gameHandler) SetGameSettings(c echo.Context) error {
	gameName := c.Param("game")

	if len(strings.TrimSpace(gameName)) == 0 || !h.service.GameExists(gameName) {
		return c.JSON(http.StatusBadRequest, &domain.Response{
			Success: false,
			Message: "Invalid game",
		})
	}

	// Validate request body
	var body params.SetGameSettingsParams
	if err := c.Bind(&body); err != nil {
		return err
	}

	if ok, err := api.ValidateRequestBody(body); !ok {
		return err
	}

	// Set game settings
	gs := &domain.GameSettings{
		BanCommandPattern: body.BanCommandPattern,
	}

	// Get the game
	game, err := h.service.GetGame(gameName)
	if err != nil {
		return err
	}

	if err := h.service.SetGameSettings(game, gs); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Game settings set",
		Payload: gs,
	})
}

func (h *gameHandler) GetDefaultGameSettings(c echo.Context) error {
	gameName := c.Param("game")

	if len(strings.TrimSpace(gameName)) == 0 || !h.service.GameExists(gameName) {
		return c.JSON(http.StatusBadRequest, &domain.Response{
			Success: false,
			Message: "Invalid game",
		})
	}

	game, err := h.service.GetGame(gameName)
	if err != nil {
		return err
	}

	defSettings := game.GetDefaultSettings()

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Payload: defSettings,
	})
}
