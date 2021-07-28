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
)

type userHandler struct {
	service     domain.UserService
	authService domain.AuthService
	authorizer  domain.Authorizer
	logger      *zap.Logger
}

func ApplyUserHandler(apiGroup *echo.Group, s domain.UserService, as domain.AuthService, a domain.Authorizer,
	protect echo.MiddlewareFunc, log *zap.Logger) {

	handler := &userHandler{
		service:     s,
		authService: as,
		authorizer:  a,
		logger:      log,
	}

	// Create the routing group
	userGroup := apiGroup.Group("/users", protect)

	// Create an enforcer to authorize users on the various endpoints
	enforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type: domain.AuthObjRefractor,
	}, log)

	// Map routes to handlers
	userGroup.GET("/", handler.GetAllUsers, enforcer.CheckAuth(authcheckers.RequireAdmin))
	userGroup.GET("/me", handler.GetOwnInfo)
	userGroup.POST("/", handler.CreateUser, enforcer.CheckAuth(authcheckers.RequireAdmin))
}

func (h *userHandler) GetAllUsers(c echo.Context) error {
	ctx := c.Request().Context()

	users, err := h.service.GetAllUsers(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: fmt.Sprintf("Fetched %d users", len(users)),
		Payload: users,
	})
}

func (h *userHandler) GetOwnInfo(c echo.Context) error {
	ctx := c.Request().Context()

	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	userInfo, err := h.service.GetByID(ctx, user.Identity.Id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Own user info fetched",
		Payload: userInfo,
	})
}

func (h *userHandler) CreateUser(c echo.Context) error {
	// Validate request body
	var body params.CreateUserParams
	if err := c.Bind(&body); err != nil {
		return err
	}

	if ok, err := api.ValidateRequestBody(body); !ok {
		return err
	}

	// Get user
	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	newUser, err := h.authService.CreateUser(c.Request().Context(), &domain.Traits{
		Email:    body.Email,
		Username: body.Username,
	}, user.Traits.Username)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &domain.Response{
		Success: true,
		Message: "User created",
		Payload: newUser,
	})
}
