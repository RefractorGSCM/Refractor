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
	"Refractor/pkg/aeshelper"
	"Refractor/pkg/querybuilders/psqlqb"
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strings"
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

	query := `INSERT INTO ChatMessages (PlayerID, Platform, ServerID, Message, Flagged, MessageVectors)
			VALUES ($1, $2, $3, $4, $5, to_tsvector($6)) RETURNING MessageID;`

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

	row := stmt.QueryRowContext(ctx, msg.PlayerID, msg.Platform, msg.ServerID, msg.Message, msg.Flagged, msg.Message)
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

	query := `SELECT MessageID, PlayerID, Platform, ServerID, Message, Flagged, CreatedAt, ModifiedAt
				FROM ChatMessages WHERE MessageID = $1;`

	results, err := r.fetch(ctx, query, id)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		return results[0], nil
	}

	return nil, errors.Wrap(domain.ErrNotFound, op)
}

func (r *chatRepo) GetRecentByServer(ctx context.Context, serverID int64, count int) ([]*domain.ChatMessage, error) {
	const op = opTag + "GetRecentByServer"

	query := `SELECT MessageID, PlayerID, Platform, ServerID, Message, Flagged, CreatedAt, ModifiedAt
			FROM ChatMessages WHERE ServerID = $1 ORDER BY CreatedAt DESC LIMIT $2;`

	results, err := r.fetch(ctx, query, serverID, count)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		return results, nil
	}

	return nil, errors.Wrap(domain.ErrNotFound, op)
}

func (r *chatRepo) Search(ctx context.Context, args domain.FindArgs, limit, offset int) (int, []*domain.ChatMessage, error) {
	const op = opTag + "Search"

	query := `
		SELECT
			cm.MessageID,
		    cm.PlayerID,
		    cm.Platform,
		    cm.ServerID,
		    cm.Message,
		    cm.Flagged,
		    cm.CreatedAt,
		    cm.ModifiedAt
		FROM ChatMessages cm
		JOIN Servers s ON s.ServerID = cm.ServerID
		WHERE
			($1::VARCHAR IS NULL OR cm.PlayerID = $1) AND
			($2::VARCHAR IS NULL OR cm.Platform = $2) AND
			($3::INT IS NULL OR cm.ServerID = $3) AND
			($4::VARCHAR IS NULL OR s.Game = $4) AND
			(($5::BIGINT IS NULL OR $6::BIGINT IS NULL) OR cm.CreatedAt BETWEEN TO_TIMESTAMP($5) AND TO_TIMESTAMP($6)) AND
			($7::VARCHAR IS NULL OR cm.MessageVectors @@ PLAINTO_TSQUERY($7))
		LIMIT $8 OFFSET $9;
	`

	var (
		playerID    = args["PlayerID"]
		platform    = args["Platform"]
		game        = args["Game"]
		serverID    = args["ServerID"]
		startDate   = args["StartDate"]
		endDate     = args["EndDate"]
		searchQuery = args["Query"]
	)

	var toQueryMethod = "PLAINTO_TSQUERY"
	searchStrPtr, ok := searchQuery.(*string)
	if ok && searchStrPtr != nil {
		var searchStr = *searchStrPtr

		if searchQuery != "" && strings.Split(searchStr, " ")[0] == "query:" {
			searchStr = searchStr[len("query: "):]
			toQueryMethod = "TO_TSQUERY"
		}

		searchQuery = searchStr
	}

	//results, err := r.fetch(ctx, query, playerID, playerID, platform, platform, serverID, serverID, game, game,
	//	startDate, endDate, startDate, endDate, searchQuery, searchQuery, limit, offset)
	results, err := r.fetch(ctx, query, playerID, platform, serverID, game, startDate, endDate, searchQuery, limit, offset)
	if err != nil {
		if strings.Contains(errors.Cause(err).Error(), "syntax error in tsquery") {
			return 0, nil, errors.Wrap(domain.ErrInvalidQuery, op)
		}

		return 0, nil, errors.Wrap(err, op)
	}

	if len(results) == 0 {
		return 0, []*domain.ChatMessage{}, nil
	}

	// Get total results count
	query = fmt.Sprintf(`
		SELECT
			COUNT(1) AS Count
		FROM ChatMessages cm
		JOIN Servers s ON s.ServerID = cm.ServerID
		WHERE
			($1::VARCHAR IS NULL OR cm.PlayerID = $1) AND
			($2::VARCHAR IS NULL OR cm.Platform = $2) AND
			($3::INT IS NULL OR cm.ServerID = $3) AND
		    ($4::VARCHAR IS NULL OR s.Game = $4) AND
			(($5::BIGINT IS NULL OR $6::BIGINT IS NULL) OR cm.CreatedAt BETWEEN TO_TIMESTAMP($5) AND TO_TIMESTAMP($6)) AND
			($7::VARCHAR IS NULL OR cm.MessageVectors @@ %s($7));
	`, toQueryMethod)

	row := r.db.QueryRowContext(ctx, query, playerID, platform, serverID, game, startDate, endDate, searchQuery)

	var resultCount int
	if err := row.Scan(&resultCount); err != nil {
		r.logger.Error("Could not get total result count while searching chat messages",
			zap.Error(err))
		return 0, nil, errors.Wrap(err, op)
	}

	return resultCount, results, err
}

func (r *chatRepo) GetFlaggedMessages(ctx context.Context, count int, serverIDs []int64) ([]*domain.ChatMessage, error) {
	const op = opTag + "GetFlaggedMessages"

	query := `
		SELECT
			MessageID,
		    PlayerID,
		    Platform,
		    ServerID,
		    Message,
		    Flagged,
		    CreatedAt,
		    ModifiedAt
		FROM ChatMessages WHERE Flagged = TRUE AND ServerID = ANY ($1::int[])
		ORDER BY CreatedAt DESC LIMIT $2;
	`

	results, err := r.fetch(ctx, query, pq.Array(serverIDs), count)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return results, nil
}

func (r *chatRepo) GetFlaggedMessageCount(ctx context.Context) (int, error) {
	const op = opTag + "GetFlaggedMessages"

	query := "SELECT COUNT(1) FROM ChatMessages WHERE Flagged = TRUE;"

	row := r.db.QueryRowContext(ctx, query)

	var count int
	if err := row.Scan(&count); err != nil {
		r.logger.Error("Could not scan flagged message count", zap.Error(err))
		return 0, errors.Wrap(err, op)
	}

	return count, nil
}

func (r *chatRepo) Update(ctx context.Context, id int64, args domain.UpdateArgs) (*domain.ChatMessage, error) {
	const op = opTag + "Update"

	query, values := r.qb.BuildUpdateQuery("ChatMessages", id, "MessageID", args)

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, values...)

	updatedMessage := &domain.ChatMessage{}
	if err := r.scanRow(row, updatedMessage); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(domain.ErrNotFound, op)
		}

		r.logger.Error("Could not scan updated chat message", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	return updatedMessage, nil
}

// Scan helpers
func (r *chatRepo) scanRow(row *sql.Row, msg *domain.ChatMessage) error {
	return row.Scan(&msg.MessageID, &msg.PlayerID, &msg.Platform, &msg.ServerID, &msg.Message, &msg.Flagged, &msg.CreatedAt, &msg.ModifiedAt)
}

func (r *chatRepo) scanRows(rows *sql.Rows, msg *domain.ChatMessage) error {
	return rows.Scan(&msg.MessageID, &msg.PlayerID, &msg.Platform, &msg.ServerID, &msg.Message, &msg.Flagged, &msg.CreatedAt, &msg.ModifiedAt)
}
