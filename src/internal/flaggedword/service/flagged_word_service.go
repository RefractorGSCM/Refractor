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
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strings"
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

	flaggedWords, err := s.repo.GetAll(ctx)
	if err != nil {
		if errors.Cause(err) == domain.ErrNotFound {
			return []*domain.FlaggedWord{}, nil
		}

		return nil, err
	}

	return flaggedWords, nil
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

func (s *flaggedWordService) MessageContainsFlaggedWord(c context.Context, message string) (bool, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	flaggedWords, err := s.repo.GetAll(ctx)
	if err != nil {
		if errors.Cause(err) == domain.ErrNotFound {
			return false, nil
		}

		return false, err
	}

	words := strings.Split(message, " ")

	flagged := false
	for _, word := range words {
		// Check if flagged words contains this word
		for _, fword := range flaggedWords {
			if strings.ToLower(word) == strings.ToLower(fword.Word) {
				flagged = true
				break
			}
		}
	}

	return flagged, nil
}
