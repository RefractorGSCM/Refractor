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
	authRepo   domain.AuthRepo
	groupRepo  domain.GroupRepo
	authorizer domain.Authorizer
	timeout    time.Duration
	logger     *zap.Logger
}

func NewUserService(ar domain.AuthRepo, gr domain.GroupRepo, a domain.Authorizer, to time.Duration, log *zap.Logger) domain.UserService {
	return &userService{
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
		newUser := &domain.User{
			ID:       au.Identity.Id,
			Username: au.Traits.Username,
		}

		// Use the authorizer to get user the user's computed permissions, scoped to Refractor.
		scope := domain.AuthScope{Type: domain.AuthObjRefractor}
		permissions, err := s.authorizer.GetPermissions(ctx, scope, newUser.ID)
		if err != nil {
			s.logger.Error("Could not get computed permissions for user", zap.String("userID", newUser.ID))
			return nil, errors.Wrap(err, fmt.Sprintf("user ID: %s", newUser.ID))
		}

		newUser.Permissions = permissions.String()

		// Use the groups repo to get the user's groups
		groups, err := s.groupRepo.GetUserGroups(ctx, newUser.ID)
		if err != nil {
			s.logger.Error("Could not get groups for user", zap.String("userID", newUser.ID))
			return nil, errors.Wrap(err, fmt.Sprintf("user ID: %s", newUser.ID))
		}

		newUser.Groups = groups

		// Add user to list
		users = append(users, newUser)
	}

	return users, nil
}

func (s *userService) CreateUser(c context.Context, traits *domain.Traits) error {
	panic("implement me")
}
