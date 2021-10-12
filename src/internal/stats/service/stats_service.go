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

type statsService struct {
	repo     domain.StatsRepo
	chatRepo domain.ChatRepo
	timeout  time.Duration
}

func NewStatsService(repo domain.StatsRepo, cr domain.ChatRepo, to time.Duration) domain.StatsService {
	return &statsService{
		repo:     repo,
		chatRepo: cr,
		timeout:  to,
	}
}

func (s *statsService) GetStats(c context.Context) (*domain.Stats, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	stats := &domain.Stats{}
	var err error

	stats.TotalPlayers, err = s.repo.GetTotalPlayers(ctx)
	if err != nil {
		return nil, err
	}

	stats.TotalInfractions, err = s.repo.GetTotalInfractions(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	oneDayAgo := now.Add(-24 * time.Hour).UTC()

	stats.NewPlayersLastDay, err = s.repo.GetTotalNewPlayersInRange(ctx, oneDayAgo, now)
	if err != nil {
		return nil, err
	}

	stats.NewInfractionsLastDay, err = s.repo.GetTotalNewInfractionsInRange(ctx, oneDayAgo, now)
	if err != nil {
		return nil, err
	}

	stats.UniquePlayersLastDay, err = s.repo.GetUniquePlayersInRange(ctx, oneDayAgo, now)
	if err != nil {
		return nil, err
	}

	stats.TotalChatMessages, err = s.repo.GetTotalChatMessages(ctx)
	if err != nil {
		return nil, err
	}

	stats.TotalFlaggedChatMessages, err = s.chatRepo.GetFlaggedMessageCount(ctx)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
