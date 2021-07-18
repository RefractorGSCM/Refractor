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
	"Refractor/params"
	"Refractor/pkg/api"
	"Refractor/pkg/bitperms"
	"Refractor/pkg/perms"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

type groupHandler struct {
	service    domain.GroupService
	authorizer domain.Authorizer
}

func ApplyGroupHandler(apiGroup *echo.Group, s domain.GroupService, authorizer domain.Authorizer, protect echo.MiddlewareFunc) {
	handler := &groupHandler{
		service:    s,
		authorizer: authorizer,
	}

	// Create the server routing group
	groupGroup := apiGroup.Group("/groups", protect)

	groupGroup.POST("/", handler.CreateGroup)
	groupGroup.GET("/", handler.GetGroups)
	groupGroup.GET("/permissions", handler.GetPermissions)
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
		Message: "permissions fetched",
		Payload: resPerms,
	})
}

func (h *groupHandler) CreateGroup(c echo.Context) error {
	// Check if the user has permission to create this groups
	ctx := c.Request().Context()
	user := c.Get("user").(*domain.AuthUser)

	authScope := domain.AuthScope{Type: domain.AuthObjRefractor}
	hasPermission, err := api.CheckPermissions(ctx, h.authorizer, authScope, user.Identity.Id, createGroupAuthChecker)
	if err != nil {
		return err
	}

	if !hasPermission {
		return c.JSON(http.StatusUnauthorized, domain.ResponseUnauthorized)
	}

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
		Message: "group created",
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

func createGroupAuthChecker(permissions *bitperms.Permissions) (bool, error) {
	if permissions.CheckFlag(perms.GetFlag(perms.FlagSuperAdmin)) {
		return true, nil
	}

	return false, nil
}
