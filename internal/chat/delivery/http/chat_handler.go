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
	"Refractor/pkg/perms"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type chatHandler struct {
	service domain.ChatService
	logger  *zap.Logger
}

func ApplyChatHandler(apiGroup *echo.Group, service domain.ChatService,
	authorizer domain.Authorizer, mware domain.Middleware, log *zap.Logger) {
	handler := &chatHandler{
		service: service,
		logger:  log,
	}

	// Create the chat routing group
	chatGroup := apiGroup.Group("/chat", mware.ProtectMiddleware, mware.ActivationMiddleware)

	// Create a server enforcer to authorize the user on the various chat endpoints
	sEnforcer := middleware.NewEnforcer(authorizer, domain.AuthScope{
		Type:        domain.AuthObjServer,
		IDFieldName: "serverId",
	}, log)

	chatGroup.GET("/recent/:serverId", handler.GetRecentServerMessages,
		sEnforcer.CheckAuth(authcheckers.HasOneOfPermissions(true, perms.FlagReadLiveChat, perms.FlagViewChatRecords)))
}

const defaultRecentMessagesCount = 20
const maxRecentMessagesCount = 100

func (h *chatHandler) GetRecentServerMessages(c echo.Context) error {
	serverIDString := c.Param("serverId")

	serverID, err := strconv.ParseInt(serverIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid server id"), http.StatusBadRequest, "")
	}

	var count int64 = defaultRecentMessagesCount
	countString := c.QueryParam("count")
	if countString != "" {
		count, err = strconv.ParseInt(countString, 10, 32)
		if err != nil {
			return &domain.HTTPError{
				Success:          false,
				Message:          "count input error",
				ValidationErrors: map[string]string{"count": "invalid int"},
				Status:           http.StatusBadRequest,
			}
		}
	}

	if count < 1 || count > maxRecentMessagesCount {
		return &domain.HTTPError{
			Success:          false,
			Message:          "count input error",
			ValidationErrors: map[string]string{"count": "should be between 1 and 100"},
			Status:           http.StatusBadRequest,
		}
	}

	ctx := c.Request().Context()
	messages, err := h.service.GetRecentByServer(ctx, serverID, int(count))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Payload: messages,
	})
}
