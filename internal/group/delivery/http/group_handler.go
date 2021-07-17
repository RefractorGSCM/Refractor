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
	"Refractor/pkg/perms"
	"fmt"
	"github.com/labstack/echo/v4"
	"math"
	"net/http"
	"time"
)

type groupHandler struct {
	service domain.GroupService
}

func ApplyGroupHandler(apiGroup *echo.Group, s domain.GroupService, authorizer domain.Authorizer, protect echo.MiddlewareFunc) {
	handler := &groupHandler{
		service: s,
	}

	// Create the server routing group
	groupGroup := apiGroup.Group("/groups", protect)

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

func (h *groupHandler) GetGroups(c echo.Context) error {
	groups := []*domain.Group{
		{
			ID:          2,
			Name:        "Super Admin",
			Color:       0xff0000,
			Position:    1,
			Permissions: "4",
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
		},
		{
			ID:          3,
			Name:        "Admin",
			Color:       0xff4d00,
			Position:    2,
			Permissions: "4",
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
		},
		{
			ID:          4,
			Name:        "Moderator",
			Color:       0x1ceb23,
			Position:    3,
			Permissions: "4",
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
		},
		{
			ID:          1,
			Name:        "Everyone",
			Color:       0xe3e3e3,
			Position:    math.MaxInt32,
			Permissions: "2",
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
		},
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: fmt.Sprintf("fetched %d groups", len(groups)),
		Payload: groups,
	})
}
