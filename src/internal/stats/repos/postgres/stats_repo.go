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

package postgres

import (
	"Refractor/domain"
	"context"
	"database/sql"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

const opTag = "StatsRepo.Postgres."

type statsRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewStatsRepo(db *sql.DB, log *zap.Logger) domain.StatsRepo {
	return &statsRepo{
		db:     db,
		logger: log,
	}
}

func (r *statsRepo) fetchCount(ctx context.Context, query string, args ...interface{}) (int, error) {
	const op = opTag + "fetchCount"

	row := r.db.QueryRowContext(ctx, query, args...)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

func (r *statsRepo) GetTotalPlayers(ctx context.Context) (int, error) {
	const op = opTag + "GetTotalPlayers"

	query := "SELECT COUNT(1) FROM Players;"

	count, err := r.fetchCount(ctx, query)
	if err != nil {
		r.logger.Error("Could not get total player count", zap.Error(err))
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

func (r *statsRepo) GetTotalInfractions(ctx context.Context) (int, error) {
	const op = opTag + "GetTotalInfractions"

	query := "SELECT COUNT(1) FROM Infractions;"

	count, err := r.fetchCount(ctx, query)
	if err != nil {
		r.logger.Error("Could not get total infraction count", zap.Error(err))
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

func (r *statsRepo) GetTotalNewPlayersInRange(ctx context.Context, start, end time.Time) (int, error) {
	const op = opTag + "GetTotalNewPlayersInRange"

	query := "SELECT COUNT(1) FROM Players WHERE CreatedAt BETWEEN $1::TIMESTAMP AND $2::TIMESTAMP;"

	count, err := r.fetchCount(ctx, query, pq.FormatTimestamp(start), pq.FormatTimestamp(end))
	if err != nil {
		r.logger.Error("Could not get new players in range count", zap.Error(err))
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

func (r *statsRepo) GetTotalNewInfractionsInRange(ctx context.Context, start, end time.Time) (int, error) {
	const op = opTag + "GetTotalNewPlayersInRange"

	query := "SELECT COUNT(1) FROM Infractions WHERE CreatedAt BETWEEN $1::TIMESTAMP AND $2::TIMESTAMP;"

	count, err := r.fetchCount(ctx, query, pq.FormatTimestamp(start), pq.FormatTimestamp(end))
	if err != nil {
		r.logger.Error("Could not get new infractions in range count", zap.Error(err))
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

func (r *statsRepo) GetUniquePlayersInRange(ctx context.Context, start, end time.Time) (int, error) {
	const op = opTag + "GetUniquePlayersInRange"

	query := "SELECT COUNT(1) FROM Players WHERE LastSeen BETWEEN $1 AND $2;"

	count, err := r.fetchCount(ctx, query, start, end)

	if err != nil {
		r.logger.Error("Could not get total total players in range count", zap.Error(err))
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

func (r *statsRepo) GetTotalChatMessages(ctx context.Context) (int, error) {
	const op = opTag + "GetTotalChatMessages"

	query := "SELECT COUNT(1) FROM ChatMessages;"

	count, err := r.fetchCount(ctx, query)
	if err != nil {
		r.logger.Error("Could not get total chat messages count", zap.Error(err))
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

func (r *statsRepo) GetTotalChatMessagesInRange(ctx context.Context, start, end time.Time) (int, error) {
	const op = opTag + "GetTotalChatMessagesInRange"

	query := "SELECT COUNT(1) FROM ChatMessages WHERE CreatedAt BETWEEN $1 AND $2;"

	count, err := r.fetchCount(ctx, query, pq.FormatTimestamp(start), pq.FormatTimestamp(end))
	if err != nil {
		r.logger.Error("Could not get new infractions in range count", zap.Error(err))
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}
