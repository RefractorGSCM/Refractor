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
	"Refractor/pkg/perms"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type chatHandler struct {
	service            domain.ChatService
	flaggedWordService domain.FlaggedWordService
	logger             *zap.Logger
}

func ApplyChatHandler(apiGroup *echo.Group, service domain.ChatService,
	fws domain.FlaggedWordService, authorizer domain.Authorizer, mware domain.Middleware, log *zap.Logger) {
	handler := &chatHandler{
		service:            service,
		flaggedWordService: fws,
		logger:             log,
	}

	// Create the chat routing group
	chatGroup := apiGroup.Group("/chat", mware.ProtectMiddleware, mware.ActivationMiddleware)

	// Create an app enforcer to authorize the user on the various chat endpoints
	rEnforcer := middleware.NewEnforcer(authorizer, domain.AuthScope{
		Type: domain.AuthObjRefractor,
	}, log)

	// Create a server enforcer to authorize the user on the various chat endpoints
	sEnforcer := middleware.NewEnforcer(authorizer, domain.AuthScope{
		Type:        domain.AuthObjServer,
		IDFieldName: "serverId",
	}, log)

	chatGroup.GET("/recent/:serverId", handler.GetRecentServerMessages,
		sEnforcer.CheckAuth(authcheckers.HasOneOfPermissions(true, perms.FlagReadLiveChat, perms.FlagViewChatRecords)))
	chatGroup.GET("/flagged", handler.GetAllFlaggedWords, rEnforcer.CheckAuth(authcheckers.RequireAdmin))
	chatGroup.POST("/flagged", handler.CreateFlaggedWord, rEnforcer.CheckAuth(authcheckers.RequireAdmin))
	chatGroup.PATCH("/flagged/:id", handler.UpdateFlaggedWord, rEnforcer.CheckAuth(authcheckers.RequireAdmin))
	chatGroup.DELETE("/flagged/:id", handler.DeleteFlaggedWord, rEnforcer.CheckAuth(authcheckers.RequireAdmin))
	chatGroup.GET("/recent/flagged", handler.GetRecentFlaggedMessages,
		rEnforcer.CheckAuth(authcheckers.HasPermission(perms.FlagModerateFlaggedMessages, true)))
	chatGroup.PATCH("/unflag/:id", handler.UnflagMessage,
		rEnforcer.CheckAuth(authcheckers.HasPermission(perms.FlagModerateFlaggedMessages, true)))
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

func (h *chatHandler) GetAllFlaggedWords(c echo.Context) error {
	allFlaggedWords, err := h.flaggedWordService.GetAll(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Payload: allFlaggedWords,
	})
}

func (h *chatHandler) CreateFlaggedWord(c echo.Context) error {
	// Validate request body
	var body params.CreateFlaggedWordParams
	if err := c.Bind(&body); err != nil {
		return err
	}

	if ok, err := api.ValidateRequestBody(body); !ok {
		return err
	}

	newWord := &domain.FlaggedWord{
		Word: body.Word,
	}

	if err := h.flaggedWordService.Store(c.Request().Context(), newWord); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &domain.Response{
		Success: true,
		Message: "Created",
		Payload: newWord,
	})
}

func (h *chatHandler) UpdateFlaggedWord(c echo.Context) error {
	flaggedWordIDString := c.Param("id")

	flaggedWordID, err := strconv.ParseInt(flaggedWordIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid flagged word id"), http.StatusBadRequest, "")
	}

	// Validate request body
	var body params.UpdateFlaggedWordParams
	if err := c.Bind(&body); err != nil {
		return err
	}

	if ok, err := api.ValidateRequestBody(body); !ok {
		return err
	}

	updated, err := h.flaggedWordService.Update(c.Request().Context(), flaggedWordID, *body.Word)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &domain.Response{
		Success: true,
		Message: "Updated",
		Payload: updated,
	})
}

func (h *chatHandler) DeleteFlaggedWord(c echo.Context) error {
	flaggedWordIDString := c.Param("id")

	flaggedWordID, err := strconv.ParseInt(flaggedWordIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid flagged word id"), http.StatusBadRequest, "")
	}

	if err := h.flaggedWordService.Delete(c.Request().Context(), flaggedWordID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Deleted",
	})
}

func (h *chatHandler) GetRecentFlaggedMessages(c echo.Context) error {
	var count int64 = 10
	countString := c.QueryParam("count")
	if countString != "" {
		var err error
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

	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	ctx := context.WithValue(c.Request().Context(), "user", user)
	messages, err := h.service.GetFlaggedMessages(ctx, int(count))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Payload: messages,
	})
}

func (h *chatHandler) UnflagMessage(c echo.Context) error {
	messageIDString := c.Param("id")

	messageID, err := strconv.ParseInt(messageIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid message id"), http.StatusBadRequest, "")
	}

	if err := h.service.UnflagMessage(c.Request().Context(), messageID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Message unflagged",
	})
}
