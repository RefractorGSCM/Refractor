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
	"github.com/guregu/null"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type infractionHandler struct {
	service           domain.InfractionService
	attachmentService domain.AttachmentService
	authorizer        domain.Authorizer
	logger            *zap.Logger
}

func ApplyInfractionHandler(apiGroup *echo.Group, s domain.InfractionService, as domain.AttachmentService,
	a domain.Authorizer, mware domain.Middleware, log *zap.Logger) {
	handler := &infractionHandler{
		service:           s,
		attachmentService: as,
		authorizer:        a,
		logger:            log,
	}

	// Create the infraction routing group
	infractionGroup := apiGroup.Group("/infractions", mware.ProtectMiddleware, mware.ActivationMiddleware)

	// Create an enforcer to authorize the user on the various infraction endpoints
	sEnforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type:        domain.AuthObjServer,
		IDFieldName: "serverId",
	}, log)

	rEnforcer := middleware.NewEnforcer(a, domain.AuthScope{
		Type: domain.AuthObjRefractor,
	}, log)

	infractionGroup.POST("/warning/:serverId", handler.CreateWarning,
		sEnforcer.CheckAuth(authcheckers.HasPermission(perms.FlagCreateWarning, true)))
	infractionGroup.POST("/mute/:serverId", handler.CreateMute,
		sEnforcer.CheckAuth(authcheckers.HasPermission(perms.FlagCreateMute, true)))
	infractionGroup.POST("/kick/:serverId", handler.CreateKick,
		sEnforcer.CheckAuth(authcheckers.HasPermission(perms.FlagCreateKick, true)))
	infractionGroup.POST("/ban/:serverId", handler.CreateBan,
		sEnforcer.CheckAuth(authcheckers.HasPermission(perms.FlagCreateBan, true)))
	infractionGroup.PATCH("/:id", handler.UpdateInfraction)              // perms checked in service
	infractionGroup.POST("/:id/repealed", handler.SetInfractionRepealed) // perms checked in service
	infractionGroup.DELETE("/:id", handler.DeleteInfraction)             // perms checked in service
	infractionGroup.GET("/player/:platform/:playerId", handler.GetPlayerInfractions,
		rEnforcer.CheckAuth(authcheckers.HasPermission(perms.FlagViewPlayerRecords, true))) // additional server specific perms checks done in service
	infractionGroup.GET("/:id", handler.GetByID)                        // perms checked in service
	infractionGroup.POST("/:id/attachment", handler.AddAttachment)      // perms checked in service
	infractionGroup.DELETE("/attachment/:id", handler.RemoveAttachment) // perms checked in service
}

type infractionRes struct {
	Attachments        []*domain.Attachment  `json:"attachments"`
	LinkedChatMessages []*domain.ChatMessage `json:"linked_chat_messages"`
	*domain.Infraction
}

func (h *infractionHandler) GetByID(c echo.Context) error {
	infractionIDString := c.Param("id")

	infractionID, err := strconv.ParseInt(infractionIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid infraction id"), http.StatusBadRequest, "")
	}

	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, "user", user)
	infraction, err := h.service.GetByID(ctx, infractionID)
	if err != nil {
		return err
	}

	// Get attachments belonging to infraction
	attachments, err := h.attachmentService.GetByInfraction(ctx, infraction.InfractionID)
	if err != nil && errors.Cause(err) != domain.ErrNotFound {
		return err
	}

	// Get chat messages linked to this infraction
	chatMessages, err := h.service.GetLinkedChatMessages(ctx, infraction.InfractionID)
	if err != nil && errors.Cause(err) != domain.ErrNotFound {
		return err
	}

	res := &infractionRes{
		LinkedChatMessages: chatMessages,
		Attachments:        attachments,
		Infraction:         infraction,
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Payload: res,
	})
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

	// Convert attachments body field to slice of attachment slices
	var attachments []*domain.Attachment
	for _, att := range body.Attachments {
		attachments = append(attachments, &domain.Attachment{
			URL:  att.URL,
			Note: att.Note,
		})
	}

	// Attach user to request context for use in the service
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, "user", user)
	newWarning, err = h.service.Store(ctx, newWarning, attachments, body.LinkedMessages)
	if err != nil {
		return err
	}

	h.logger.Info("Warning record created",
		zap.Int64("Infraction ID", newWarning.InfractionID),
		zap.String("Player ID", newWarning.PlayerID),
		zap.String("Platform", newWarning.Platform),
		zap.Int64("Server ID", newWarning.ServerID),
		zap.String("User ID", newWarning.UserID.ValueOrZero()),
		zap.String("Reason", newWarning.Reason.ValueOrZero()),
	)

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
		Duration:     null.NewInt(int64(*body.Duration), true),
		SystemAction: false,
		CreatedAt:    null.Time{},
		ModifiedAt:   null.Time{},
	}

	// Convert attachments body field to slice of attachment slices
	var attachments []*domain.Attachment
	for _, att := range body.Attachments {
		attachments = append(attachments, &domain.Attachment{
			URL:  att.URL,
			Note: att.Note,
		})
	}

	// Attach user to request context for use in the service
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, "user", user)
	newMute, err = h.service.Store(ctx, newMute, attachments, body.LinkedMessages)
	if err != nil {
		return err
	}

	h.logger.Info("Mute record created",
		zap.Int64("Infraction ID", newMute.InfractionID),
		zap.String("Player ID", newMute.PlayerID),
		zap.String("Platform", newMute.Platform),
		zap.Int64("Server ID", newMute.ServerID),
		zap.String("User ID", newMute.UserID.ValueOrZero()),
		zap.String("Reason", newMute.Reason.ValueOrZero()),
		zap.Int64("Duration", newMute.Duration.ValueOrZero()),
	)

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

	// Convert attachments body field to slice of attachment slices
	var attachments []*domain.Attachment
	for _, att := range body.Attachments {
		attachments = append(attachments, &domain.Attachment{
			URL:  att.URL,
			Note: att.Note,
		})
	}

	// Attach user to request context for use in the service
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, "user", user)
	newKick, err = h.service.Store(ctx, newKick, attachments, body.LinkedMessages)
	if err != nil {
		return err
	}

	h.logger.Info("Kick record created",
		zap.Int64("Infraction ID", newKick.InfractionID),
		zap.String("Player ID", newKick.PlayerID),
		zap.String("Platform", newKick.Platform),
		zap.Int64("Server ID", newKick.ServerID),
		zap.String("User ID", newKick.UserID.ValueOrZero()),
		zap.String("Reason", newKick.Reason.ValueOrZero()),
	)

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
		Duration:     null.NewInt(int64(*body.Duration), true),
		SystemAction: false,
		CreatedAt:    null.Time{},
		ModifiedAt:   null.Time{},
	}

	// Convert attachments body field to slice of attachment slices
	var attachments []*domain.Attachment
	for _, att := range body.Attachments {
		attachments = append(attachments, &domain.Attachment{
			URL:  att.URL,
			Note: att.Note,
		})
	}

	// Attach user to request context for use in the service
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, "user", user)
	newBan, err = h.service.Store(ctx, newBan, attachments, body.LinkedMessages)
	if err != nil {
		return err
	}

	h.logger.Info("Ban record created",
		zap.Int64("Infraction ID", newBan.InfractionID),
		zap.String("Player ID", newBan.PlayerID),
		zap.String("Platform", newBan.Platform),
		zap.Int64("Server ID", newBan.ServerID),
		zap.String("User ID", newBan.UserID.ValueOrZero()),
		zap.String("Reason", newBan.Reason.ValueOrZero()),
		zap.Int64("Duration", newBan.Duration.ValueOrZero()),
	)

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

	// Attach user to request context for use in the service
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, "user", user)

	// Update warning
	updated, err := h.service.Update(ctx, infractionID, updateArgs)
	if err != nil {
		return err
	}

	h.logger.Info("Infraction updated",
		zap.Any("Update Args", updateArgs),
		zap.String("User ID", user.Identity.Id),
	)

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Infraction updated",
		Payload: updated,
	})
}

func (h *infractionHandler) SetInfractionRepealed(c echo.Context) error {
	infractionIDString := c.Param("id")

	infractionID, err := strconv.ParseInt(infractionIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid infraction id"), http.StatusBadRequest, "")
	}

	// Validate request body
	var body params.SetInfractionRepealedParams
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

	// Attach user to request context for use in the service
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, "user", user)

	// Set repealed status of infraction
	updated, err := h.service.SetRepealed(ctx, infractionID, body.Repealed)
	if err != nil {
		return err
	}

	h.logger.Info("Infraction repeal status set",
		zap.Any("Is Repealed", updated.Repealed),
		zap.String("User ID", user.Identity.Id),
	)

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Infraction repeal status set",
		Payload: updated,
	})
}

func (h *infractionHandler) DeleteInfraction(c echo.Context) error {
	infractionIDString := c.Param("id")

	infractionID, err := strconv.ParseInt(infractionIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid infraction id"), http.StatusBadRequest, "")
	}

	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	// Attach user to request context for use in the service
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, "user", user)

	// Delete the infraction
	if err := h.service.Delete(ctx, infractionID); err != nil {
		return err
	}

	h.logger.Info("Infraction deleted",
		zap.Int64("Infraction ID", infractionID),
		zap.String("User ID", user.Identity.Id),
	)

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Infraction deleted",
	})
}

func (h *infractionHandler) GetPlayerInfractions(c echo.Context) error {
	platform := c.Param("platform")
	playerID := c.Param("playerId")

	validPlatform := false
	for _, p := range domain.AllPlatforms {
		if platform == p {
			validPlatform = true
			break
		}
	}

	if !validPlatform {
		return c.JSON(http.StatusBadRequest, &domain.Response{
			Success: false,
			Message: "Invalid platform provided",
		})
	}

	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	// Get request context and attach the user to it so that the service.GetByPlayer call is authorized against the
	// various servers.
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, "user", user)

	// Get player infractions
	infractions, err := h.service.GetByPlayer(ctx, playerID, platform)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: fmt.Sprintf("Fetched %d infractions", len(infractions)),
		Payload: infractions,
	})
}

func (h *infractionHandler) AddAttachment(c echo.Context) error {
	infractionIDString := c.Param("id")

	infractionID, err := strconv.ParseInt(infractionIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid infraction id"), http.StatusBadRequest, "")
	}

	// Validate request body
	var body params.CreateAttachmentParams
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

	// Attach user to request context for use in the service
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, "user", user)

	// Create attachment
	attachment := &domain.Attachment{
		InfractionID: infractionID,
		URL:          body.URL,
		Note:         body.Note,
	}

	if err := h.attachmentService.Store(ctx, attachment); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Attachment created",
		Payload: attachment,
	})
}

// RemoveAttachment would belong more in a separate Attachment handler, but it doesn't hurt anything while being here.
// If something else calls for the creation of a separate attachment service, this method should be moved there.
func (h *infractionHandler) RemoveAttachment(c echo.Context) error {
	attachmentIDString := c.Param("id")

	attachmentID, err := strconv.ParseInt(attachmentIDString, 10, 64)
	if err != nil {
		return domain.NewHTTPError(fmt.Errorf("invalid attachment id"), http.StatusBadRequest, "")
	}

	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not cast user to *domain.AuthUser")
	}

	// Attach user to request context for use in the service
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, "user", user)

	// Delete attachment
	if err := h.attachmentService.Delete(ctx, attachmentID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: "Attachment deleted",
	})
}
