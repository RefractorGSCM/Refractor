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
	"go.uber.org/zap"
	"time"
)

type flaggedWordService struct {
	repo    domain.FlaggedWordRepo
	timeout time.Duration
	logger  *zap.Logger
}

func NewFlaggedWordService(repo domain.FlaggedWordRepo, to time.Duration, log *zap.Logger) domain.FlaggedWordService {
	return &flaggedWordService{
		repo:    repo,
		timeout: to,
		logger:  log,
	}
}

func (s *flaggedWordService) Store(c context.Context, word *domain.FlaggedWord) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Store(ctx, word)
}

func (s *flaggedWordService) GetAll(c context.Context) ([]*domain.FlaggedWord, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.GetAll(ctx)
}

func (s *flaggedWordService) Update(c context.Context, id int64, newWord string) (*domain.FlaggedWord, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Update(ctx, id, newWord)
}

func (s *flaggedWordService) Delete(c context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.repo.Delete(ctx, id)
}
