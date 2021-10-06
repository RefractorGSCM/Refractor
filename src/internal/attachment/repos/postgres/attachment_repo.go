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

const opTag = "AttachmentRepo.Postgres."

type attachmentRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewAttachmentRepo(db *sql.DB, logger *zap.Logger) domain.AttachmentRepo {
	return &attachmentRepo{
		db:     db,
		logger: logger,
	}
}

func (r *attachmentRepo) fetch(ctx context.Context, query string, args ...interface{}) ([]*domain.Attachment, error) {
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

	results := make([]*domain.Attachment, 0)
	for rows.Next() {
		attachment := &domain.Attachment{}

		if err := r.scanRows(rows, attachment); err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.Wrap(domain.ErrNotFound, op)
			}

			return nil, errors.Wrap(err, op)
		}

		results = append(results, attachment)
	}

	return results, nil
}

func (r *attachmentRepo) Store(ctx context.Context, attachment *domain.Attachment) error {
	const op = opTag + "Store"

	query := "INSERT INTO Attachments (InfractionID, URL, Note) VALUES ($1, $2, $3) RETURNING AttachmentID;"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, attachment.InfractionID, attachment.URL, attachment.Note)

	var id int64
	if err := row.Scan(&id); err != nil {
		r.logger.Error("Could not execute prepared statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	attachment.AttachmentID = id

	return nil
}

func (r *attachmentRepo) GetByInfraction(ctx context.Context, infractionID int64) ([]*domain.Attachment, error) {
	const op = opTag + "GetByInfraction"

	query := "SELECT * FROM Attachments WHERE InfractionID = $1;"

	results, err := r.fetch(ctx, query, infractionID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		return results, nil
	}

	return nil, errors.Wrap(domain.ErrNotFound, op)
}

func (r *attachmentRepo) GetByID(ctx context.Context, id int64) (*domain.Attachment, error) {
	const op = opTag + "GetByID"

	query := "SELECT * FROM Attachments WHERE AttachmentID = $1;"

	results, err := r.fetch(ctx, query, id)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		return results[0], nil
	}

	return nil, errors.Wrap(domain.ErrNotFound, op)
}

func (r *attachmentRepo) Delete(ctx context.Context, id int64) error {
	const op = opTag + "Delete"

	query := "DELETE FROM Attachments WHERE AttachmentID = $1;"

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
func (r *attachmentRepo) scanRow(row *sql.Row, attachment *domain.Attachment) error {
	return row.Scan(&attachment.AttachmentID, &attachment.InfractionID, &attachment.URL, &attachment.Note)
}

func (r *attachmentRepo) scanRows(rows *sql.Rows, attachment *domain.Attachment) error {
	return rows.Scan(&attachment.AttachmentID, &attachment.InfractionID, &attachment.URL, &attachment.Note)
}
