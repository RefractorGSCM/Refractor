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
	"Refractor/pkg/querybuilders/psqlqb"
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const opTag = "InfractionRepo.Postgres."

type infractionRepo struct {
	db     *sql.DB
	logger *zap.Logger
	qb     domain.QueryBuilder
}

func NewInfractionRepo(db *sql.DB, logger *zap.Logger) domain.InfractionRepo {
	return &infractionRepo{
		db:     db,
		logger: logger,
		qb:     psqlqb.NewPostgresQueryBuilder(),
	}
}

func (r *infractionRepo) fetch(ctx context.Context, query string, args ...interface{}) ([]*domain.Infraction, error) {
	const op = opTag + "Fetch"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Could not execute SQL query", zap.String("query", query), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	// Clean up on function exit
	defer func() {
		errRow := rows.Close()
		if errRow != nil {
			r.logger.Warn("Could not close SQL rows", zap.Error(err))
		}
	}()

	results := make([]*domain.Infraction, 0)
	for rows.Next() {
		infraction := &domain.DBInfraction{}

		if err := r.scanRows(rows, infraction); err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.Wrap(domain.ErrNotFound, op)
			}

			return nil, errors.Wrap(err, op)
		}

		results = append(results, infraction.Infraction())
	}

	return results, nil
}

func (r *infractionRepo) Store(ctx context.Context, i *domain.DBInfraction) (*domain.Infraction, error) {
	const op = opTag + "Store"

	query := `INSERT INTO Infractions(PlayerID, Platform, UserID, ServerID, Type, Reason, Duration, SystemAction)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, i.PlayerID, i.Platform, i.UserID, i.ServerID, i.Type, i.Reason, i.Duration, i.SystemAction)

	dbi := &domain.DBInfraction{}

	if err := r.scanRow(row, dbi); err != nil {
		r.logger.Error("Could not scan newly created infraction", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	return dbi.Infraction(), nil
}

func (r *infractionRepo) GetByID(ctx context.Context, id int64) (*domain.Infraction, error) {
	panic("implement me")
}

func (r *infractionRepo) Update(ctx context.Context, id int64, args domain.UpdateArgs) (*domain.Infraction, error) {
	panic("implement me")
}

func (r *infractionRepo) Delete(ctx context.Context, id int64) error {
	panic("implement me")
}

// Scan helpers
func (r *infractionRepo) scanRow(row *sql.Row, i *domain.DBInfraction) error {
	return row.Scan(&i.InfractionID, &i.PlayerID, &i.Platform, &i.UserID, &i.ServerID, &i.Type, &i.Reason, &i.Duration, &i.SystemAction, &i.CreatedAt, &i.ModifiedAt)
}

func (r *infractionRepo) scanRows(rows *sql.Rows, i *domain.DBInfraction) error {
	return rows.Scan(&i.InfractionID, &i.PlayerID, &i.Platform, &i.UserID, &i.ServerID, &i.Type, &i.Reason, &i.Duration, &i.SystemAction, &i.CreatedAt, &i.ModifiedAt)
}
