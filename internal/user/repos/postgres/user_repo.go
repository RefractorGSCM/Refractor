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

const opTag = "UserRepo.Postgres."

type userRepo struct {
	db     *sql.DB
	logger *zap.Logger
	qb     domain.QueryBuilder
}

func NewUserRepo(db *sql.DB, log *zap.Logger) domain.UserMetaRepo {
	return &userRepo{
		db:     db,
		logger: log,
		qb:     psqlqb.NewPostgresQueryBuilder(),
	}
}

func (r *userRepo) fetch(ctx context.Context, query string, args ...interface{}) ([]*domain.UserMeta, error) {
	const op = opTag + "fetch"

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

	results := make([]*domain.UserMeta, 0)
	for rows.Next() {
		meta := &domain.UserMeta{}

		if err := r.scanRows(rows, meta); err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.Wrap(domain.ErrNotFound, op)
			}

			return nil, errors.Wrap(err, op)
		}

		results = append(results, meta)
	}

	return results, nil
}

func (r *userRepo) Store(ctx context.Context, meta *domain.UserMeta) error {
	const op = opTag + "Store"

	query := "INSERT INTO UserMeta (UserID, InitialUsername, Username, Deactivated) VALUES ($1, $2, $3, $4);"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	_, err = stmt.ExecContext(ctx, meta.ID, meta.InitialUsername, meta.Username, meta.Deactivated)
	if err != nil {
		r.logger.Error("Could not execute statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *userRepo) GetByID(ctx context.Context, userID string) (*domain.UserMeta, error) {
	const op = opTag + "GetByID"

	query := "SELECT * FROM UserMeta WHERE UserID = $1;"

	results, err := r.fetch(ctx, query, userID)
	if err != nil {
		r.logger.Error("Could not get user by id", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		return results[0], nil
	}

	return nil, errors.Wrap(domain.ErrNotFound, op)
}

func (r *userRepo) Update(ctx context.Context, id string, args domain.UpdateArgs) (*domain.UserMeta, error) {
	const op = opTag + "Update"

	query, values := r.qb.BuildUpdateQuery("UserMeta", id, "UserID", args)

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, values...)

	updatedMeta := &domain.UserMeta{}
	if err := r.scanRow(row, updatedMeta); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(domain.ErrNotFound, op)
		}

		r.logger.Error("Could not scan updated meta", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	return updatedMeta, nil
}

// Scan helpers
func (r *userRepo) scanRow(row *sql.Row, meta *domain.UserMeta) error {
	return row.Scan(&meta.ID, &meta.InitialUsername, &meta.Username, &meta.Deactivated)
}

func (r *userRepo) scanRows(rows *sql.Rows, meta *domain.UserMeta) error {
	return rows.Scan(&meta.ID, &meta.InitialUsername, &meta.Username, &meta.Deactivated)
}
