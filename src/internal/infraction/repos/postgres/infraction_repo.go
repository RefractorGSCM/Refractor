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
	"github.com/lib/pq"
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
		infraction := &domain.Infraction{}

		if err := r.scanRows(rows, infraction); err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.Wrap(domain.ErrNotFound, op)
			}

			return nil, errors.Wrap(err, op)
		}

		results = append(results, infraction)
	}

	return results, nil
}

func (r *infractionRepo) Store(ctx context.Context, i *domain.Infraction) (*domain.Infraction, error) {
	const op = opTag + "Store"

	query := `INSERT INTO Infractions (PlayerID, Platform, UserID, ServerID, Type, Reason, Duration, SystemAction)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, i.PlayerID, i.Platform, i.UserID, i.ServerID, i.Type, i.Reason, i.Duration, i.SystemAction)

	infraction := &domain.Infraction{}

	if err := r.scanRow(row, infraction); err != nil {
		r.logger.Error("Could not scan newly created infraction", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	return infraction, nil
}

func (r *infractionRepo) GetByID(ctx context.Context, id int64) (*domain.Infraction, error) {
	const op = opTag + "GetByID"

	query := "SELECT * FROM Infractions WHERE InfractionID = $1;"

	results, err := r.fetch(ctx, query, id)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		return results[0], nil
	}

	return nil, errors.Wrap(domain.ErrNotFound, op)
}

func (r *infractionRepo) GetByPlayer(ctx context.Context, playerID, platform string) ([]*domain.Infraction, error) {
	const op = opTag + "GetByPlayer"

	query := "SELECT * FROM Infractions WHERE PlayerID = $1 AND Platform = $2;"

	results, err := r.fetch(ctx, query, playerID, platform)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		return results, nil
	}

	return nil, errors.Wrap(domain.ErrNotFound, op)
}

func (r *infractionRepo) Update(ctx context.Context, id int64, args domain.UpdateArgs) (*domain.Infraction, error) {
	const op = opTag + "Update"

	query, values := r.qb.BuildUpdateQuery("Infractions", id, "InfractionID", args)

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, values...)

	updatedInfraction := &domain.Infraction{}
	if err := r.scanRow(row, updatedInfraction); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(domain.ErrNotFound, op)
		}

		r.logger.Error("Could not scan updated infraction", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	return updatedInfraction, nil
}

func (r *infractionRepo) Delete(ctx context.Context, id int64) error {
	const op = opTag + "Delete"

	query := "DELETE FROM Infractions WHERE InfractionID = $1;"

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

func (r *infractionRepo) Search(ctx context.Context, args domain.FindArgs, limit, offset int) (int, []*domain.Infraction, error) {
	const op = opTag + "Search"

	query := `
		SELECT
			res.*,
			um.Username AS StaffName
		FROM (
		    SELECT
		    	i.*
			FROM Infractions i
			INNER JOIN Servers s ON i.ServerID = s.ServerID
			WHERE
				($1::VARCHAR IS NULL OR i.Type = $2) AND
				($3::VARCHAR IS NULL OR i.PlayerID = $4) AND
				($5::VARCHAR IS NULL OR i.Platform = $6) AND
				($7::VARCHAR IS NULL OR i.UserID = $8) AND
				($9::INT IS NULL OR i.ServerID = $10) AND
				($11::VARCHAR IS NULL OR s.Game = $12)
			) res
		JOIN UserMeta um ON res.UserID = um.UserID
		LIMIT $13 OFFSET $14;
	`

	var (
		iType    = args["Type"]
		playerID = args["PlayerID"]
		platform = args["Platform"]
		userID   = args["UserID"]
		serverID = args["ServerID"]
		game     = args["Game"]
	)

	rows, err := r.db.QueryContext(ctx, query, iType, iType, playerID, playerID, platform, platform, userID, userID,
		serverID, serverID, game, game, limit, offset)
	if err != nil {
		r.logger.Error("Could not execute infraction search query",
			zap.Any("Filters", args),
			zap.Error(err),
		)
		return 0, nil, errors.Wrap(err, op)
	}

	var results []*domain.Infraction

	for rows.Next() {
		res := &domain.Infraction{}

		if err := rows.Scan(&res.InfractionID, &res.PlayerID, &res.Platform, &res.UserID, &res.ServerID, &res.Type,
			&res.Reason, &res.Duration, &res.SystemAction, &res.CreatedAt, &res.ModifiedAt, &res.IssuerName); err != nil {
			r.logger.Error("Could not scan infraction search result", zap.Error(err))
			return 0, nil, errors.Wrap(err, op)
		}

		results = append(results, res)
	}

	if len(results) < 1 {
		return 0, []*domain.Infraction{}, nil
	}

	// Get total number of matches
	query = `
		SELECT COUNT(1) AS Count
		FROM Infractions i
		INNER JOIN Servers s ON i.ServerID = s.ServerID
		WHERE
			($1::VARCHAR IS NULL OR i.Type = $2) AND
			($3::VARCHAR IS NULL OR i.PlayerID = $4) AND
		    ($5::VARCHAR IS NULL OR i.Platform = $6) AND
			($7::VARCHAR IS NULL OR i.UserID = $8) AND
			($9::INT IS NULL OR i.ServerID = $10) AND
			($11::VARCHAR IS NULL OR s.Game = $12)
	`

	row := r.db.QueryRowContext(ctx, query, iType, iType, playerID, playerID, platform, platform, userID, userID, serverID, serverID,
		game, game)

	var count int
	if err := row.Scan(&count); err != nil {
		r.logger.Error("Could not scan total search results for infraction search", zap.Error(err))
		return 0, nil, errors.Wrap(err, op)
	}

	return count, results, nil
}

func (r *infractionRepo) GetLinkedChatMessages(ctx context.Context, id int64) ([]*domain.ChatMessage, error) {
	const op = opTag + "GetLinkedChatMessages"

	// This query makes the infraction repo dependent on the chat messages table being present in postgres.
	// This could be problematic if we ever switch to storing chat messages elsewhere, but for now this is the
	// most efficient way of doing it, both in terms of development time and performance.
	//
	// If this becomes a problem in the future, this method could instead be made to return a slice of IDs
	// which can then be used to fetch messages from another repo.

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
		FROM InfractionChatMessages icm
		INNER JOIN ChatMessages cm on cm.MessageID = icm.MessageID
		WHERE icm.InfractionID = 1;
	`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Could not execute linked infraction-chatmessages query", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	messages := make([]*domain.ChatMessage, 0)
	for rows.Next() {
		msg := &domain.ChatMessage{}

		if err := rows.Scan(&msg.MessageID, &msg.PlayerID, &msg.Platform, &msg.ServerID, &msg.Message,
			&msg.Flagged, &msg.CreatedAt, &msg.ModifiedAt); err != nil {
			r.logger.Error("Could not scan chat message", zap.Error(err))
			return nil, errors.Wrap(err, op)
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

const pgForeignKeyErrCode = pq.ErrorCode("23503")

func (r *infractionRepo) LinkChatMessage(ctx context.Context, id int64, messageID int64) error {
	const op = opTag + "LinkChatMessage"

	query := "INSERT INTO InfractionChatMessages (InfractionID, MessageID) VALUES ($1, $2);"

	_, err := r.db.ExecContext(ctx, query, id, messageID)
	if err != nil {
		if pgErr, ok := err.(pq.Error); ok {
			if pgErr.Code == pgForeignKeyErrCode {
				return errors.Wrap(domain.ErrNotFound, op)
			}
		}

		r.logger.Error("Could not link chat message to infraction",
			zap.Int64("Infraction ID", id),
			zap.Int64("Message ID", messageID),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (r *infractionRepo) UnlinkChatMessage(ctx context.Context, id int64, messageID int64) error {
	const op = opTag + "UnlinkChatMessage"

	query := "DELETE FROM InfractionChatMessages WHERE InfractionID = $1 AND MessageID = $2;"

	res, err := r.db.ExecContext(ctx, query, id, messageID)
	if err != nil {
		r.logger.Error("Could not unlink chat message to infraction",
			zap.Int64("Infraction ID", id),
			zap.Int64("Message ID", messageID),
			zap.Error(err),
		)
		return err
	}

	if rows, err := res.RowsAffected(); err != nil {
		r.logger.Error("Could not get affected rows", zap.Error(err))
		return errors.Wrap(err, op)
	} else if rows < 1 {
		return errors.Wrap(domain.ErrNotFound, op)
	}

	return nil
}

// Scan helpers
func (r *infractionRepo) scanRow(row *sql.Row, i *domain.Infraction) error {
	return row.Scan(&i.InfractionID, &i.PlayerID, &i.Platform, &i.UserID, &i.ServerID, &i.Type, &i.Reason, &i.Duration, &i.SystemAction, &i.CreatedAt, &i.ModifiedAt)
}

func (r *infractionRepo) scanRows(rows *sql.Rows, i *domain.Infraction) error {
	return rows.Scan(&i.InfractionID, &i.PlayerID, &i.Platform, &i.UserID, &i.ServerID, &i.Type, &i.Reason, &i.Duration, &i.SystemAction, &i.CreatedAt, &i.ModifiedAt)
}
