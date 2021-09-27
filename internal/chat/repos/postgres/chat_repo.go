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

const opTag = "ChatRepo.Postgres."

type chatRepo struct {
	db     *sql.DB
	logger *zap.Logger
	qb     domain.QueryBuilder
}

func NewChatRepo(db *sql.DB, logger *zap.Logger) domain.ChatRepo {
	return &chatRepo{
		db:     db,
		logger: logger,
		qb:     psqlqb.NewPostgresQueryBuilder(),
	}
}

func (r *chatRepo) fetch(ctx context.Context, query string, args ...interface{}) ([]*domain.ChatMessage, error) {
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

	results := make([]*domain.ChatMessage, 0)
	for rows.Next() {
		msg := &domain.ChatMessage{}

		if err := r.scanRows(rows, msg); err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.Wrap(domain.ErrNotFound, op)
			}

			return nil, errors.Wrap(err, op)
		}

		results = append(results, msg)
	}

	return results, nil
}

// Store stores a new chat message in the postgres database. The following fields must be present on the passed in
// chat message struct:
//
// PlayerID, Platform, ServerID, Message
//
// Flagged is optional.
func (r *chatRepo) Store(ctx context.Context, msg *domain.ChatMessage) error {
	const op = opTag + "Store"

	query := `INSERT INTO ChatMessages (PlayerID, Platform, ServerID, Message, Flagged)
			VALUES ($1, $2, $3, $4, $5) RETURNING MessageID;`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Could not begin ChatMessage store transaction", zap.Error(err))
		return errors.Wrap(err, op)
	}

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		_ = tx.Rollback()
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, msg.PlayerID, msg.Platform, msg.ServerID, msg.Message, msg.Flagged)
	if err != nil {
		_ = tx.Rollback()
		r.logger.Error("Could not execute query", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		_ = tx.Rollback()
		r.logger.Error("Could not scan inserted ID", zap.Error(err))
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("Could not commit ChatMessage store transaction", zap.Error(err))
		return errors.Wrap(err, op)
	}

	msg.MessageID = id
	return nil
}

func (r *chatRepo) GetByID(ctx context.Context, id int64) (*domain.ChatMessage, error) {
	const op = opTag + "GetByID"

	query := `SELECT * FROM ChatMessages WHERE MessageID = $1;`

	results, err := r.fetch(ctx, query, id)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		return results[0], nil
	}

	return nil, errors.Wrap(domain.ErrNotFound, op)
}

// Scan helpers
func (r *chatRepo) scanRow(row *sql.Row, msg *domain.ChatMessage) error {
	return row.Scan(&msg.MessageID, &msg.PlayerID, &msg.Platform, &msg.ServerID, &msg.Message, &msg.Flagged, &msg.CreatedAt, &msg.ModifiedAt)
}

func (r *chatRepo) scanRows(rows *sql.Rows, msg *domain.ChatMessage) error {
	return rows.Scan(&msg.MessageID, &msg.PlayerID, &msg.Platform, &msg.ServerID, &msg.Message, &msg.Flagged, &msg.CreatedAt, &msg.ModifiedAt)
}
