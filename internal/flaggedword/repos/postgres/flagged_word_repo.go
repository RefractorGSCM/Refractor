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
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const opTag = "FlaggedWordRepo.Postgres."

type repo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewFlaggedWordRepo(db *sql.DB, log *zap.Logger) domain.FlaggedWordRepo {
	return &repo{
		db:     db,
		logger: log,
	}
}

func (r *repo) fetch(ctx context.Context, query string, args ...interface{}) ([]*domain.FlaggedWord, error) {
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

	results := make([]*domain.FlaggedWord, 0)
	for rows.Next() {
		res := &domain.FlaggedWord{}

		if err := r.scanRows(rows, res); err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.Wrap(domain.ErrNotFound, op)
			}

			return nil, errors.Wrap(err, op)
		}

		results = append(results, res)
	}

	return results, nil
}

func (r *repo) Store(ctx context.Context, word *domain.FlaggedWord) error {
	const op = opTag + "Store"

	query := `INSERT INTO FlaggedWords (Word) VALUES ($1) RETURNING WordID;`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, word.Word)

	var id int64
	if err := row.Scan(&id); err != nil {
		r.logger.Error("Could not scan inserted flagged word ID", zap.Error(err))
		return errors.Wrap(err, op)
	}

	word.ID = id

	return nil
}

func (r *repo) GetAll(ctx context.Context) ([]*domain.FlaggedWord, error) {
	const op = opTag + "GetAll"

	query := "SELECT * FROM FlaggedWords;"

	results, err := r.fetch(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		return results, nil
	}

	return nil, errors.Wrap(domain.ErrNotFound, op)
}

func (r *repo) Update(ctx context.Context, id int64, newWord string) (*domain.FlaggedWord, error) {
	const op = opTag + "Update"

	query := "UPDATE FlaggedWords SET Word = $1 WHERE WordID = $2 RETURNING *;"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, id)

	updated := &domain.FlaggedWord{}
	if err := r.scanRow(row, updated); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(domain.ErrNotFound, op)
		}

		r.logger.Error("Could not scan updated flagged word", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	return updated, nil
}

func (r *repo) Delete(ctx context.Context, id int64) error {
	const op = opTag + "Delete"

	query := "DELETE FROM FlaggedWords WHERE WordID = $1;"

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Could not execute query", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Could not get affected rows", zap.Error(err))
		return errors.Wrap(err, op)
	}

	if rowsAffected < 1 {
		return errors.Wrap(domain.ErrNotFound, op)
	}

	return nil
}

// Scan helpers
func (r *repo) scanRow(row *sql.Row, fw *domain.FlaggedWord) error {
	return row.Scan(&fw.ID, &fw.Word)
}

func (r *repo) scanRows(rows *sql.Rows, fw *domain.FlaggedWord) error {
	return rows.Scan(&fw.ID, &fw.Word)
}
