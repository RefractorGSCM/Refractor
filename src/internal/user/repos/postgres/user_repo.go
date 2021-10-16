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
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

const opTag = "UserRepo.Postgres."

type userRepo struct {
	db     *sql.DB
	logger *zap.Logger
	qb     domain.QueryBuilder
	cache  *gocache.Cache
}

func NewUserRepo(db *sql.DB, log *zap.Logger) domain.UserMetaRepo {
	return &userRepo{
		db:     db,
		logger: log,
		qb:     psqlqb.NewPostgresQueryBuilder(),
		cache:  gocache.New(time.Hour*1, time.Hour*1),
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

	// If the user meta is cached, pull it from cache and return it.
	cachedUser, isCached := r.cache.Get(userID)
	if isCached {
		foundUser := cachedUser.(*domain.UserMeta)
		return foundUser, nil
	}

	// Otherwise, fetch it from the database and then cache it
	query := "SELECT * FROM UserMeta WHERE UserID = $1;"

	results, err := r.fetch(ctx, query, userID)
	if err != nil {
		r.logger.Error("Could not get user by id", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		result := results[0]

		r.cache.SetDefault(userID, result)

		return result, nil
	}

	return nil, errors.Wrap(domain.ErrNotFound, op)
}

func (r *userRepo) Update(ctx context.Context, userID string, args domain.UpdateArgs) (*domain.UserMeta, error) {
	const op = opTag + "Update"

	query, values := r.qb.BuildUpdateQuery("UserMeta", userID, "UserID", args, nil)

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

	// Update cache
	r.cache.SetDefault(userID, updatedMeta)

	return updatedMeta, nil
}

func (r *userRepo) IsDeactivated(ctx context.Context, userID string) (bool, error) {
	const op = opTag + "IsDeactivated"

	query := "SELECT EXISTS(SELECT 1 FROM UserMeta WHERE Deactivated = TRUE AND UserID = $1);"

	isDeactivated := false
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&isDeactivated); err != nil {
		r.logger.Error("Could not scan row", zap.Error(err))
		return false, errors.Wrap(err, op)
	}

	return isDeactivated, nil
}

func (r *userRepo) GetUsername(ctx context.Context, userID string) (string, error) {
	const op = opTag + "IsDeactivated"

	// If the user meta is cached, pull it from cache and return it.
	cachedUser, isCached := r.cache.Get(userID)
	if isCached {
		foundUser := cachedUser.(*domain.UserMeta)
		return foundUser.Username, nil
	}

	query := "SELECT Username FROM UserMeta WHERE UserID = $1;"

	var username string
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&username); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return "", errors.Wrap(domain.ErrNotFound, op)
		}

		r.logger.Error("Could not scan row", zap.Error(err))
		return "", errors.Wrap(err, op)
	}

	return username, nil
}

const pgUniqueViolationCode = "23505"

func (r *userRepo) LinkPlayer(ctx context.Context, userID, platform, playerID string) error {
	const op = opTag + "LinkPlayer"

	query := "INSERT INTO UserPlayers (UserID, Platform, PlayerID) VALUES ($1, $2, $3);"

	_, err := r.db.ExecContext(ctx, query, userID, platform, playerID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgUniqueViolationCode {
			// if this player is already linked to an account, we expect a unique violation error
			return errors.Wrap(domain.ErrConflict, op)
		}

		r.logger.Error("Could not insert into UserPlayers table",
			zap.String("User ID", userID),
			zap.String("Platform", platform),
			zap.String("Player ID", playerID),
			zap.Error(err))
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *userRepo) UnlinkPlayer(ctx context.Context, userID, platform, playerID string) error {
	const op = opTag + "UnlinkPlayer"

	query := "DELETE FROM UserPlayers WHERE UserID = $1 AND Platform = $2 AND PlayerID = $3;"

	res, err := r.db.ExecContext(ctx, query, userID, platform, playerID)
	if err != nil {
		r.logger.Error("Could not delete from UserPlayers table",
			zap.String("User ID", userID),
			zap.String("Platform", platform),
			zap.String("Player ID", playerID),
			zap.Error(err))
		return errors.Wrap(err, op)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		r.logger.Error("Could not get rows affected from UserPlayers table delete",
			zap.Error(err))
		return errors.Wrap(err, op)
	} else if rowsAffected < 1 {
		return errors.Wrap(domain.ErrNotFound, op)
	}

	return nil
}

func (r *userRepo) GetLinkedPlayers(ctx context.Context, userID string) ([]*domain.Player, error) {
	const op = opTag + "GetLinkedPlayers"

	query := `
		SELECT
			p.*
		FROM UserPlayers up
		INNER JOIN Players p ON p.Platform = up.Platform AND p.PlayerID = up.PlayerID
		WHERE up.UserID = $1;
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Could not query UserPlayers table",
			zap.String("User ID", userID),
			zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	results := make([]*domain.Player, 0)
	for rows.Next() {
		res := &domain.Player{}

		if err := rows.Scan(&res.PlayerID, &res.Platform, &res.Watched, &res.LastSeen,
			&res.CreatedAt, &res.ModifiedAt); err != nil {
			r.logger.Error("Could not scan player result", zap.Error(err))
			return nil, errors.Wrap(err, op)
		}

		results = append(results, res)
	}

	return results, nil
}

// Scan helpers
func (r *userRepo) scanRow(row *sql.Row, meta *domain.UserMeta) error {
	return row.Scan(&meta.ID, &meta.InitialUsername, &meta.Username, &meta.Deactivated)
}

func (r *userRepo) scanRows(rows *sql.Rows, meta *domain.UserMeta) error {
	return rows.Scan(&meta.ID, &meta.InitialUsername, &meta.Username, &meta.Deactivated)
}
