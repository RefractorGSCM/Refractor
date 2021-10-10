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

package domain

import (
	"context"
	"time"
)

type Stats struct {
	TotalPlayers         int `json:"total_players"`
	TotalInfractions     int `json:"total_infractions"`
	NewPlayersLastDay    int `json:"new_players_last_day"`
	UniquePlayersLastDay int `json:"unique_players_last_day"`
}

type StatsRepo interface {
	GetTotalPlayers(ctx context.Context) (int, error)
	GetTotalInfractions(ctx context.Context) (int, error)
	GetTotalNewPlayersInRange(ctx context.Context, start, end time.Time) (int, error)
	GetUniquePlayersInRange(ctx context.Context, start, end time.Time) (int, error)
}

type StatsService interface {
	GetStats(c context.Context) (*Stats, error)
}
