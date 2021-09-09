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
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type attachmentHandler struct {
	service    domain.AttachmentService
	authorizer domain.Authorizer
	logger     *zap.Logger
}

func ApplyAttachmentHandler(apiGroup *echo.Group, s domain.AttachmentService, a domain.Authorizer, mware domain.Middleware, log *zap.Logger) {
	//handler := &attachmentHandler{
	//	service: s,
	//	authorizer: a,
	//	logger: log,
	//}
	//
	//// Create the routing group
	//attachmentGroup := apiGroup.Group("/attachments", mware.ProtectMiddleware)
	//
	//enforcer := middleware.NewEnforcer(a, domain.AuthScope{
	//	Type: domain.AuthObjRefractor,
	//}, log)
}
