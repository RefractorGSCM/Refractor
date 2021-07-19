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

func ApplyGroupHandler(apiGroup *echo.Group, s domain.GroupService, a domain.Authorizer, protect echo.MiddlewareFunc, log *zap.Logger) {
	handler := &groupHandler{
		service:    s,
		authorizer: a,
		logger:     log,
	}

	// Create the server routing group
	groupGroup := apiGroup.Group("/groups", protect)

	// Create an enforcer to authorize the user on the various endpoints
	enforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type: domain.AuthObjRefractor,
	}, log)

	groupGroup.POST("/", handler.CreateGroup, enforcer.CheckAuth(authcheckers.DenyAll))
	groupGroup.GET("/", handler.GetGroups)
	groupGroup.GET("/permissions", handler.GetPermissions)
	groupGroup.DELETE("/:id", handler.DeleteGroup, enforcer.CheckAuth(authcheckers.DenyAll))
	groupGroup.PUT("/:id", handler.UpdateGroup, enforcer.CheckAuth(authcheckers.DenyAll))
}

type resPermission struct {
	ID          int            `json:"id"`
	Name        perms.FlagName `json:"name"`
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

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Group deleted",
	})
}

func (h *groupHandler) UpdateGroup(c echo.Context) error {
	// Parse targte group ID
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

	// Build update args
	args := domain.UpdateArgs{}

	if body.Name != "" {
		args["Name"] = body.Name
	}

	if body.Color != nil {
		args["Color"] = *body.Color
	}

	if body.Position != 0 {
		args["Position"] = body.Position
	}

	if body.Permissions != "" {
		args["Permissions"] = body.Permissions
	}

	if len(args) < 1 {
		return c.JSON(http.StatusBadRequest, &domain.Response{
			Success: false,
			Message: "No update fields provided",
		})
	}

	// Update
	updated, err := h.service.Update(c.Request().Context(), groupID, args)
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
