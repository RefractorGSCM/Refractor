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
	"time"
)

type groupService struct {
	repo    domain.GroupRepo
	timeout time.Duration
}

func NewGroupService(repo domain.GroupRepo, timeout time.Duration) domain.GroupService {
	return &groupService{
		repo:    repo,
		timeout: timeout,
	}
}

func (s *groupService) Store(c context.Context, group *domain.Group) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Store(ctx, group)
}

func (s *groupService) GetAll(c context.Context) ([]*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	groups, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Add base group to the results
	baseGroup, err := s.repo.GetBaseGroup(ctx)
	if err != nil {
		return nil, err
	}
	groups = append(groups, baseGroup)

	return domain.GroupSlice(groups).SortByPosition(), nil
}

func (s *groupService) GetByID(c context.Context, id int64) (*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.GetByID(ctx, id)
}

func (s *groupService) Delete(c context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Delete(ctx, id)
}

func (s *groupService) Update(c context.Context, id int64, args domain.UpdateArgs) (*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Update(ctx, id, args)
}

func (s *groupService) Reorder(c context.Context, newPositions []*domain.GroupReorderInfo) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Reorder(ctx, newPositions)
}

func (s *groupService) UpdateBase(c context.Context, args domain.UpdateArgs) (*domain.Group, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	currentBase, err := s.repo.GetBaseGroup(ctx)
	if err != nil {
		return nil, err
	}

	// Only allow the updating of Permissions and Color
	if args["Permissions"] != nil {
		updatedPermissions := args["Permissions"].(*string)
		currentBase.Permissions = *updatedPermissions
	}

	if args["Color"] != nil {
		updatedColor := args["Color"].(*int)
		currentBase.Color = *updatedColor
	}

	// Set the base group
	if err := s.repo.SetBaseGroup(ctx, currentBase); err != nil {
		return nil, err
	}

	return currentBase, nil
}
