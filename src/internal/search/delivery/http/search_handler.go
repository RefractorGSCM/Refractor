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
	"Refractor/pkg/structutils"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

type searchHandler struct {
	service domain.SearchService
}

func ApplySearchHandler(apiGroup *echo.Group, s domain.SearchService, a domain.Authorizer,
	mware domain.Middleware, log *zap.Logger) {
	handler := &searchHandler{
		service: s,
	}

	// Create the search routing group
	searchGroup := apiGroup.Group("/search", mware.ProtectMiddleware, mware.ActivationMiddleware)

	// Create an enforcer to authorize the user on the various endpoints
	enforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type: domain.AuthObjRefractor,
	}, log)

	searchGroup.POST("/players", handler.SearchPlayers,
		enforcer.CheckAuth(authcheckers.HasPermission(perms.FlagViewPlayerRecords, true)))
	searchGroup.POST("/infractions", handler.SearchInfractions,
		enforcer.CheckAuth(authcheckers.HasPermission(perms.FlagViewInfractionRecords, true)))
	searchGroup.POST("/chat", handler.SearchChatMessages,
		enforcer.CheckAuth(authcheckers.HasPermission(perms.FlagViewChatRecords, true)))
}

type searchRes struct {
	Total   int         `json:"total"`
	Results interface{} `json:"results"`
}

func (h *searchHandler) SearchPlayers(c echo.Context) error {
	// Validate request body
	var body params.SearchPlayerParams
	if err := c.Bind(&body); err != nil {
		return err
	}

	if ok, err := api.ValidateRequestBody(body); !ok {
		return err
	}

	// Execute search
	total, results, err := h.service.SearchPlayers(c.Request().Context(), body.Term, body.Type, body.Platform, body.Limit, body.Offset)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Payload: &searchRes{
			Total:   total,
			Results: results,
		},
	})
}

func (h *searchHandler) SearchInfractions(c echo.Context) error {
	// Validate request body
	var body params.SearchInfractionParams
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

	// Get search args
	searchArgs, err := structutils.GetNonNilFieldMap(body)
	if err != nil {
		return err
	}

	if len(searchArgs) < 1 {
		return c.JSON(http.StatusBadRequest, &domain.Response{
			Success: false,
			Message: "No search fields provided",
		})
	}

	// Execute search
	ctx := context.WithValue(c.Request().Context(), "user", user)
	total, results, err := h.service.SearchInfractions(ctx, searchArgs, body.Limit, body.Offset)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Payload: &searchRes{
			Total:   total,
			Results: results,
		},
	})
}

func (h *searchHandler) SearchChatMessages(c echo.Context) error {
	// Validate request body
	var body params.SearchMessagesParams
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

	// Get search args
	searchArgs, err := structutils.GetNonNilFieldMap(body)
	if err != nil {
		return err
	}

	if len(searchArgs) < 1 {
		return c.JSON(http.StatusBadRequest, &domain.Response{
			Success: false,
			Message: "No search fields provided",
		})
	}

	// Execute search
	ctx := context.WithValue(c.Request().Context(), "user", user)
	total, results, err := h.service.SearchChatMessages(ctx, searchArgs, body.Limit, body.Offset)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Payload: &searchRes{
			Total:   total,
			Results: results,
		},
	})
}
