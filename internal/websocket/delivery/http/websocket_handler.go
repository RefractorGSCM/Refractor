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
	"Refractor/pkg/websocket"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type websocketHandler struct {
	service domain.WebsocketService
	logger  *zap.Logger
}

func ApplyWebsocketHandler(echo *echo.Echo, ws domain.WebsocketService, mware domain.Middleware, log *zap.Logger) {
	handler := &websocketHandler{
		service: ws,
		logger:  log,
	}

	echo.Any("/ws", handler.WebsocketHandler, mware.ProtectMiddleware, mware.ActivationMiddleware)
}

func (h *websocketHandler) WebsocketHandler(c echo.Context) error {
	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user as *domain.AuthUser")
	}

	conn, err := websocket.Upgrade(c.Response(), c.Request())
	if err != nil {
		h.logger.Error("Could not upgrade websocket request", zap.String("User ID", user.Identity.Id), zap.Error(err))
		return err
	}

	// Create a client for this server
	h.service.CreateClient(user.Identity.Id, conn)

	return nil
}
