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
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type groupHandler struct {
	service    domain.GroupService
	authorizer domain.Authorizer
	logger     *zap.Logger
}

func ApplyGroupHandler(apiGroup *echo.Group, s domain.GroupService, a domain.Authorizer, mware domain.Middleware, log *zap.Logger) {
	handler := &groupHandler{
		service:    s,
		authorizer: a,
		logger:     log,
	}

	// Create the routing group
	groupGroup := apiGroup.Group("/groups", mware.ProtectMiddleware)

	act := mware.ActivationMiddleware

	// Create an enforcer to authorize the user on the various endpoints
	enforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type: domain.AuthObjRefractor,
	}, log)

	groupGroup.POST("/", handler.CreateGroup, act, enforcer.CheckAuth(authcheckers.DenyAll))
	groupGroup.GET("/", handler.GetGroups, act)
	groupGroup.GET("/permissions", handler.GetPermissions)
	groupGroup.DELETE("/:id", handler.DeleteGroup, act, enforcer.CheckAuth(authcheckers.DenyAll))
	groupGroup.PATCH("/:id", handler.UpdateGroup, act, enforcer.CheckAuth(authcheckers.DenyAll))
	groupGroup.PATCH("/base", handler.UpdateBaseGroup, act, enforcer.CheckAuth(authcheckers.DenyAll))
	groupGroup.PATCH("/order", handler.ReorderGroups, act, enforcer.CheckAuth(authcheckers.DenyAll))
	groupGroup.PUT("/users/add", handler.SetUserGroup(true), act, enforcer.CheckAuth(authcheckers.RequireAdmin))
	groupGroup.PUT("/users/remove", handler.SetUserGroup(false), act, enforcer.CheckAuth(authcheckers.RequireAdmin))
	groupGroup.GET("/servers/:id", handler.GetServerOverrides, act, enforcer.CheckAuth(authcheckers.RequireAdmin))
}

type resPermission struct {
	ID          int            `json:"id"`
	Name        perms.FlagName `json:"name"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description"`
	Flag        string         `json:"flag"`
}

func (h *groupHandler) GetPermissions(c echo.Context) error {
	permissions := perms.GetAll()

	var resPerms []*resPermission

	for _, perm := range permissions {
		resPerms = append(resPerms, &resPermission{
			ID:          perm.ID,
			Name:        perm.Name,
			DisplayName: perm.DisplayName,
			Description: perms.GetDescription(perm.Name),
			Flag:        perm.Flag.String(),
		})
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Permissions fetched",
		Payload: resPerms,
	})
}

func (h *groupHandler) CreateGroup(c echo.Context) error {
	// Validate request body
	var body params.CreateGroupParams
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

	// Create the new group
	newGroup := &domain.Group{
		Name:        body.Name,
		Color:       body.Color,
		Position:    body.Position,
		Permissions: body.Permissions,
	}

	if err := h.service.Store(c.Request().Context(), newGroup); err != nil {
		return err
	}

	h.logger.Info("Group created",
		zap.Int64("Group ID", newGroup.ID),
		zap.String("Created By", user.Identity.Id),
	)

	return c.JSON(http.StatusCreated, &domain.Response{
		Success: true,
		Message: "Group created",
		Payload: newGroup,
	})
}

func (h *groupHandler) GetGroups(c echo.Context) error {
	groups, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: fmt.Sprintf("fetched %d groups", len(groups)),
		Payload: groups,
	})
}

func (h *groupHandler) DeleteGroup(c echo.Context) error {
	groupIDString := c.Param("id")

	groupID, err := strconv.ParseInt(groupIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid group id"), http.StatusBadRequest, "")
	}

	if err := h.service.Delete(c.Request().Context(), groupID); err != nil {
		return err
	}

	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	h.logger.Info("Group deleted",
		zap.Int64("Group ID", groupID),
		zap.String("Deleted By", user.Identity.Id),
	)

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Group deleted",
	})
}

func (h *groupHandler) UpdateGroup(c echo.Context) error {
	// Parse target group ID
	groupIDString := c.Param("id")

	groupID, err := strconv.ParseInt(groupIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid group id"), http.StatusBadRequest, "")
	}

	// Validate request body
	var body params.UpdateGroupParams
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

	// Update
	updated, err := h.service.Update(c.Request().Context(), groupID, updateArgs)
	if err != nil {
		return err
	}

	// Return updated group
	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Group updated",
		Payload: updated,
	})
}

func (h *groupHandler) UpdateBaseGroup(c echo.Context) error {
	// Validate request body
	var body params.UpdateGroupParams
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

	// Disalllow the updating of Name and Position for the base group
	delete(updateArgs, "Name")
	delete(updateArgs, "Position")

	if len(updateArgs) < 1 {
		return c.JSON(http.StatusBadRequest, &domain.Response{
			Success: false,
			Message: "No update fields provided",
		})
	}

	// Update base group
	updated, err := h.service.UpdateBase(c.Request().Context(), updateArgs)
	if err != nil {
		return err
	}

	// Return updated group
	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Base group updated",
		Payload: updated,
	})
}

func (h *groupHandler) ReorderGroups(c echo.Context) error {
	var body params.GroupReorderArray
	if err := c.Bind(&body); err != nil {
		return err
	}

	if ok, err := api.ValidateRequestBody(&body); !ok {
		return err
	}

	var griArr []*domain.GroupReorderInfo
	for _, gri := range body {
		griArr = append(griArr, &domain.GroupReorderInfo{
			GroupID: gri.GroupID,
			NewPos:  gri.NewPos,
		})
	}

	if err := h.service.Reorder(c.Request().Context(), griArr); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Groups reordered",
	})
}

func (h *groupHandler) SetUserGroup(add bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var body params.SetUserGroupParams
		if err := c.Bind(&body); err != nil {
			return err
		}

		if ok, err := api.ValidateRequestBody(body); !ok {
			fmt.Println(err)

			httpErr, ok := err.(*domain.HTTPError)
			if !ok {
				return fmt.Errorf("err was not an http error")
			}

			return httpErr
		}

		user, ok := c.Get("user").(*domain.AuthUser)
		if !ok {
			return fmt.Errorf("could not cast user to *domain.AuthUser")
		}

		groupSetCtx := domain.GroupSetContext{
			SetterUserID: user.Identity.Id,
			TargetUserID: body.UserID,
			GroupID:      body.GroupID,
		}

		ctx := c.Request().Context()
		if add {
			if err := h.service.AddUserGroup(ctx, groupSetCtx); err != nil {
				return err
			}

			return c.JSON(http.StatusOK, &domain.Response{
				Success: true,
				Message: "Group added",
			})
		} else {
			if err := h.service.RemoveUserGroup(ctx, groupSetCtx); err != nil {
				return err
			}

			return c.JSON(http.StatusOK, &domain.Response{
				Success: true,
				Message: "Group removed",
			})
		}
	}
}

func (h *groupHandler) GetServerOverrides(c echo.Context) error {
	// Parse target server ID
	serverIDString := c.Param("id")

	serverID, err := strconv.ParseInt(serverIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid group id"), http.StatusBadRequest, "")
	}

	overrides, err := h.service.GetServerOverridesAllGroups(c.Request().Context(), serverID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Overrides fetched",
		Payload: overrides,
	})
}
