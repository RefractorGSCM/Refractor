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

package service

import (
	"Refractor/domain"
	"Refractor/pkg/perms"
	"context"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type attachmentService struct {
	repo           domain.AttachmentRepo
	infractionRepo domain.InfractionRepo
	authorizer     domain.Authorizer
	timeout        time.Duration
	logger         *zap.Logger
}

func NewAttachmentService(repo domain.AttachmentRepo, ir domain.InfractionRepo, a domain.Authorizer, to time.Duration,
	log *zap.Logger) domain.AttachmentService {
	return &attachmentService{
		repo:           repo,
		infractionRepo: ir,
		authorizer:     a,
		timeout:        to,
		logger:         log,
	}
}

// Store stores a new attachment. If a user is set in the passed in context with the key of "user" then the user's
// authorization will be checked to see if they have permission to create a new infraction on the target infraction.
// Otherwise, this is assumed to be a system call and authorization is skipped.
func (s *attachmentService) Store(c context.Context, attachment *domain.Attachment) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// If user is set in context, check authorization to modify attachments for the target infraction.
	user, ok := ctx.Value("user").(*domain.AuthUser)
	if ok {
		hasPermission, err := s.canAttachOnInfraction(ctx, attachment.InfractionID, user)
		if err != nil {
			return err
		}

		if !hasPermission {
			return domain.NewHTTPError(nil, http.StatusUnauthorized,
				"You do not have permission to create attachments for this infraction.")
		}
	}

	return s.repo.Store(ctx, attachment)
}

func (s *attachmentService) GetByInfraction(c context.Context, infractionID int64) ([]*domain.Attachment, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.GetByInfraction(ctx, infractionID)
}

// Delete deletes an attachment. If a user is set in the passed in context with the key of "user" then the user's
// authorization will be checked to see if they have permission to delete attachments on the target infraction.
// Otherwise, this is assumed to be a system call and authorization is skipped.
func (s *attachmentService) Delete(c context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// If user is set in context, check authorization to modify attachments for the target infraction.
	user, ok := ctx.Value("user").(*domain.AuthUser)
	if ok {
		attachment, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}

		hasPermission, err := s.canAttachOnInfraction(ctx, attachment.InfractionID, user)
		if err != nil {
			return err
		}

		if !hasPermission {
			return domain.NewHTTPError(nil, http.StatusUnauthorized,
				"You do not have permission to delete attachments for this infraction.")
		}
	}

	return s.repo.Delete(ctx, id)
}

func (s *attachmentService) canAttachOnInfraction(ctx context.Context, infractionID int64, user *domain.AuthUser) (bool, error) {
	infraction, err := s.infractionRepo.GetByID(ctx, infractionID)
	if err != nil {
		return false, err
	}

	// The user will be granted permission to update this infraction if any of the following paths are satisfied:
	// 1. The user is an admin or super admin
	// OR:
	// 1. The target infraction was created by the user
	// 2. The user has permission to edit infraction records created by them.
	// OR:
	// 1. The user has permission to edit any infraction, even those not created by them.

	// Get computed user permissions for the given server
	userPerms, err := s.authorizer.GetPermissions(ctx, domain.AuthScope{
		Type: domain.AuthObjServer,
		ID:   infraction.ServerID,
	}, user.Identity.Id)
	if err != nil {
		return false, err
	}

	// Grant access if the user is an admin or super admin
	if userPerms.CheckFlag(perms.GetFlag(perms.FlagAdministrator)) || userPerms.CheckFlag(perms.GetFlag(perms.FlagSuperAdmin)) {
		return true, nil
	}

	// Grant access if the target infraction was created by the user and they have permission to edit their own infractions
	if infraction.UserID.Valid && infraction.UserID.ValueOrZero() == user.Identity.Id &&
		userPerms.CheckFlag(perms.GetFlag(perms.FlagEditOwnInfractions)) {
		return true, nil
	}

	// Grant access if the user has permission to edit any infraction
	if userPerms.CheckFlag(perms.GetFlag(perms.FlagEditAnyInfractions)) {
		return true, nil
	}

	// Otherwise, deny access
	return false, nil
}
