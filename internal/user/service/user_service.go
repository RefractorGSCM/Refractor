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
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type userService struct {
	metaRepo   domain.UserMetaRepo
	authRepo   domain.AuthRepo
	groupRepo  domain.GroupRepo
	authorizer domain.Authorizer
	timeout    time.Duration
	logger     *zap.Logger
}

func NewUserService(mr domain.UserMetaRepo, ar domain.AuthRepo, gr domain.GroupRepo, a domain.Authorizer, to time.Duration, log *zap.Logger) domain.UserService {
	return &userService{
		metaRepo:   mr,
		authRepo:   ar,
		groupRepo:  gr,
		authorizer: a,
		timeout:    to,
		logger:     log,
	}
}

func (s *userService) GetAllUsers(c context.Context) ([]*domain.User, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	authUsers, err := s.authRepo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	var users []*domain.User

	for _, au := range authUsers {
		newUser, err := s.getUserInfo(ctx, au)
		if err != nil {
			return nil, err
		}

		// Add user to list
		users = append(users, newUser)
	}

	return users, nil
}

func (s *userService) GetByID(c context.Context, userID string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	au, err := s.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Could not GetUserByID from auth repo", zap.String("UserID", userID), zap.Error(err))
		return nil, err
	}

	user, err := s.getUserInfo(ctx, au)
	if err != nil {
		s.logger.Error("Could not get user info", zap.String("UserID", userID), zap.Error(err))
		return nil, err
	}

	return user, nil
}

func (s *userService) getUserInfo(ctx context.Context, authUser *domain.AuthUser) (*domain.User, error) {
	newUser := &domain.User{
		ID:       authUser.Identity.Id,
		Username: authUser.Traits.Username,
	}

	// Use the authorizer to get user the user's computed permissions, scoped to Refractor.
	scope := domain.AuthScope{Type: domain.AuthObjRefractor}
	permissions, err := s.authorizer.GetPermissions(ctx, scope, newUser.ID)
	if err != nil {
		s.logger.Error("Could not get computed permissions for user", zap.String("userID", newUser.ID), zap.Error(err))
		return nil, errors.Wrap(err, fmt.Sprintf("user ID: %s", newUser.ID))
	}

	newUser.Permissions = permissions.String()

	// Use the groups repo to get the user's groups
	groups, err := s.groupRepo.GetUserGroups(ctx, newUser.ID)
	if errors.Cause(err) == domain.ErrNotFound {
		groups = []*domain.Group{}
	} else if err != nil {
		s.logger.Error("Could not get groups for user", zap.String("userID", newUser.ID), zap.Error(err))
		return nil, errors.Wrap(err, fmt.Sprintf("user ID: %s", newUser.ID))
	}

	newUser.Groups = groups

	// Get user meta
	meta, err := s.metaRepo.GetByID(ctx, authUser.Identity.Id)
	if err != nil {
		s.logger.Error("Could not get meta for user", zap.String("userID", newUser.ID), zap.Error(err))
		return nil, errors.Wrap(err, fmt.Sprintf("user ID: %s", newUser.ID))
	}

	newUser.UserMeta = meta

	return newUser, nil
}
