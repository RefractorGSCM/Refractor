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
	"github.com/guregu/null"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type infractionHandler struct {
	service    domain.InfractionService
	authorizer domain.Authorizer
	logger     *zap.Logger
}

func ApplyInfractionHandler(apiGroup *echo.Group, s domain.InfractionService, a domain.Authorizer, mware domain.Middleware, log *zap.Logger) {
	handler := &infractionHandler{
		service:    s,
		authorizer: a,
		logger:     log,
	}

	// Create the infraction routing group
	infractionGroup := apiGroup.Group("/infractions", mware.ProtectMiddleware, mware.ActivationMiddleware)

	// Create an enforcer to authorize the user on the various infraction endpoints
	sEnforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type:        domain.AuthObjServer,
		IDFieldName: "serverId",
	}, log)

	//rEnforcer := middleware.NewEnforcer(a, domain.AuthScope{
	//	Type: domain.AuthObjRefractor,
	//}, log)

	infractionGroup.POST("/warning/:serverId", handler.CreateWarning,
		sEnforcer.CheckAuth(authcheckers.HasPermission(perms.FlagCreateWarning, true)))
	infractionGroup.POST("/mute/:serverId", handler.CreateMute,
		sEnforcer.CheckAuth(authcheckers.HasPermission(perms.FlagCreateMute, true)))
	infractionGroup.POST("/kick/:serverId", handler.CreateKick,
		sEnforcer.CheckAuth(authcheckers.HasPermission(perms.FlagCreateKick, true)))
	infractionGroup.POST("/ban/:serverId", handler.CreateBan,
		sEnforcer.CheckAuth(authcheckers.HasPermission(perms.FlagCreateBan, true)))
	infractionGroup.PATCH("/:id", handler.UpdateInfraction)
}

func (h *infractionHandler) CreateWarning(c echo.Context) error {
	serverIDString := c.Param("serverId")

	serverID, err := strconv.ParseInt(serverIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid server id"), http.StatusBadRequest, "")
	}

	// Validate request body
	var body params.CreateWarningParams
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

	// Create the new warning
	newWarning := &domain.Infraction{
		PlayerID:     body.PlayerID,
		Platform:     body.Platform,
		UserID:       null.NewString(user.Identity.Id, true),
		ServerID:     serverID,
		Type:         domain.InfractionTypeWarning,
		Reason:       null.NewString(body.Reason, true),
		Duration:     null.Int{},
		SystemAction: false,
		CreatedAt:    null.Time{},
		ModifiedAt:   null.Time{},
	}

	newWarning, err = h.service.Store(c.Request().Context(), newWarning)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &domain.Response{
		Success: true,
		Message: "Warning created",
		Payload: newWarning,
	})
}

func (h *infractionHandler) CreateMute(c echo.Context) error {
	serverIDString := c.Param("serverId")

	serverID, err := strconv.ParseInt(serverIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid server id"), http.StatusBadRequest, "")
	}

	// Validate request body
	var body params.CreateMuteParams
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

	// Create the new warning
	newMute := &domain.Infraction{
		PlayerID:     body.PlayerID,
		Platform:     body.Platform,
		UserID:       null.NewString(user.Identity.Id, true),
		ServerID:     serverID,
		Type:         domain.InfractionTypeMute,
		Reason:       null.NewString(body.Reason, true),
		Duration:     null.NewInt(int64(body.Duration), true),
		SystemAction: false,
		CreatedAt:    null.Time{},
		ModifiedAt:   null.Time{},
	}

	newMute, err = h.service.Store(c.Request().Context(), newMute)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &domain.Response{
		Success: true,
		Message: "Mute created",
		Payload: newMute,
	})
}

func (h *infractionHandler) CreateKick(c echo.Context) error {
	serverIDString := c.Param("serverId")

	serverID, err := strconv.ParseInt(serverIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid server id"), http.StatusBadRequest, "")
	}

	// Validate request body
	var body params.CreateKickParams
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

	// Create the new warning
	newKick := &domain.Infraction{
		PlayerID:     body.PlayerID,
		Platform:     body.Platform,
		UserID:       null.NewString(user.Identity.Id, true),
		ServerID:     serverID,
		Type:         domain.InfractionTypeKick,
		Reason:       null.NewString(body.Reason, true),
		Duration:     null.Int{},
		SystemAction: false,
		CreatedAt:    null.Time{},
		ModifiedAt:   null.Time{},
	}

	newKick, err = h.service.Store(c.Request().Context(), newKick)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &domain.Response{
		Success: true,
		Message: "Kick created",
		Payload: newKick,
	})
}

func (h *infractionHandler) CreateBan(c echo.Context) error {
	serverIDString := c.Param("serverId")

	serverID, err := strconv.ParseInt(serverIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid server id"), http.StatusBadRequest, "")
	}

	// Validate request body
	var body params.CreateBanParams
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

	// Create the new warning
	newBan := &domain.Infraction{
		PlayerID:     body.PlayerID,
		Platform:     body.Platform,
		UserID:       null.NewString(user.Identity.Id, true),
		ServerID:     serverID,
		Type:         domain.InfractionTypeBan,
		Reason:       null.NewString(body.Reason, true),
		Duration:     null.NewInt(int64(body.Duration), true),
		SystemAction: false,
		CreatedAt:    null.Time{},
		ModifiedAt:   null.Time{},
	}

	newBan, err = h.service.Store(c.Request().Context(), newBan)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &domain.Response{
		Success: true,
		Message: "Ban created",
		Payload: newBan,
	})
}

func (h *infractionHandler) UpdateInfraction(c echo.Context) error {
	infractionIDString := c.Param("id")

	infractionID, err := strconv.ParseInt(infractionIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid infraction id"), http.StatusBadRequest, "")
	}

	// Validate request body
	var body params.UpdateInfractionParams
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

	// Update warning
	updated, err := h.service.Update(c.Request().Context(), infractionID, updateArgs)
	if err != nil {
		return err
	}

	h.logger.Info("Warning updated",
		zap.Any("Update Args", updateArgs),
		zap.String("User ID", user.Identity.Id),
	)

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Warning updated",
		Payload: updated,
	})
}
