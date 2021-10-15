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
	"context"
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
	mware domain.Middleware, log *zap.Logger) {

	handler := &userHandler{
		service:     s,
		authService: as,
		authorizer:  a,
		logger:      log,
	}

	// Create the routing group
	userGroup := apiGroup.Group("/users", mware.ProtectMiddleware)

	// Create an enforcer to authorize users on the various endpoints
	enforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type: domain.AuthObjRefractor,
	}, log)

	// Map routes to handlers
	userGroup.GET("/", handler.GetAllUsers, mware.ActivationMiddleware, enforcer.CheckAuth(authcheckers.RequireAdmin))
	userGroup.GET("/me", handler.GetOwnInfo)
	userGroup.POST("/", handler.CreateUser, mware.ActivationMiddleware, enforcer.CheckAuth(authcheckers.RequireAdmin))
	userGroup.PATCH("/deactivate/:id", handler.ChangeUserActivation(false), mware.ActivationMiddleware)
	userGroup.PATCH("/reactivate/:id", handler.ChangeUserActivation(true), mware.ActivationMiddleware)
	userGroup.POST("/link/player", handler.LinkPlayer, mware.ActivationMiddleware, enforcer.CheckAuth(authcheckers.RequireAdmin))
	userGroup.POST("/unlink/player", handler.UnlinkPlayer, mware.ActivationMiddleware, enforcer.CheckAuth(authcheckers.RequireAdmin))
	userGroup.GET("/link/player/:id", handler.GetLinkedPlayers, mware.ActivationMiddleware, enforcer.CheckAuth(authcheckers.RequireAdmin))
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

	ctx := c.Request().Context()
	// Get user
	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	newUser, err := h.authService.CreateUser(ctx, &domain.Traits{
		Email:    body.Email,
		Username: body.Username,
	}, user.Traits.Username)
	if err != nil {
		return err
	}

	userInfo, err := h.service.GetByID(ctx, newUser.Identity.Id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &domain.Response{
		Success: true,
		Message: "User created",
		Payload: userInfo,
	})
}

func (h *userHandler) ChangeUserActivation(shouldBeActivated bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		targetUserID := c.Param("id")

		user, ok := c.Get("user").(*domain.AuthUser)
		if !ok {
			return fmt.Errorf("could not cast user to *domain.AuthUser")
		}

		ctx := c.Request().Context()
		ctx = context.WithValue(ctx, "userids", map[string]string{
			"Setter": user.Identity.Id,
			"Target": targetUserID,
		})

		var message string

		if shouldBeActivated {
			if err := h.service.ReactivateUser(ctx, targetUserID); err != nil {
				return err
			}

			h.logger.Info("User account reactivated",
				zap.String("Reactivated UserID", targetUserID),
				zap.String("Reactivated By", user.Identity.Id))

			message = "User account reactivated"
		} else {
			if err := h.service.DeactivateUser(ctx, targetUserID); err != nil {
				return err
			}

			h.logger.Info("User account deactivated",
				zap.String("Deactivated UserID", targetUserID),
				zap.String("Deactivated By", user.Identity.Id))

			message = "User account deactivated"
		}

		return c.JSON(http.StatusOK, &domain.Response{
			Success: true,
			Message: message,
		})
	}
}

func (h *userHandler) LinkPlayer(c echo.Context) error {
	// Validate request body
	var body params.LinkPlayerParams
	if err := c.Bind(&body); err != nil {
		return err
	}

	if ok, err := api.ValidateRequestBody(body); !ok {
		return err
	}

	if err := h.service.LinkPlayer(c.Request().Context(), body.UserID, body.Platform, body.PlayerID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Player linked to user",
	})
}

func (h *userHandler) UnlinkPlayer(c echo.Context) error {
	// Validate request body
	var body params.LinkPlayerParams
	if err := c.Bind(&body); err != nil {
		return err
	}

	if ok, err := api.ValidateRequestBody(body); !ok {
		return err
	}

	if err := h.service.UnlinkPlayer(c.Request().Context(), body.UserID, body.Platform, body.PlayerID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Player unlinked",
	})
}

func (h *userHandler) GetLinkedPlayers(c echo.Context) error {
	userID := c.Param("id")

	players, err := h.service.GetLinkedPlayers(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Payload: players,
	})
}
