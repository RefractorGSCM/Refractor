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
	"Refractor/pkg/perms"
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const opTag = "GroupRepo.Postgres."

type groupRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewGroupRepo(db *sql.DB, logger *zap.Logger) (domain.GroupRepo, error) {
	repo := &groupRepo{
		db:     db,
		logger: logger,
	}

	// Check if a group with ID 1 (everyone) exists. If it does not, we create it.
	if _, err := repo.GetByID(context.TODO(), 1); errors.Cause(err) == domain.ErrNotFound {
		if err := repo.createDefaultGroup(); err != nil {
			return nil, err
		}
	}

	return repo, nil
}

func (r *groupRepo) createDefaultGroup() error {
	const op = opTag + "createDefaultGroup"

	newGroup := &domain.Group{
		ID:          1,
		Name:        "Everyone",
		Color:       0xb0b0b0,
		Position:    1,
		Permissions: perms.GetDefaultPermissions().String(),
	}

	query := "SELECT MAX(Position) FROM Groups;"

	row := r.db.QueryRow(query)

	var highestPosition *int = new(int)
	if err := row.Scan(&highestPosition); err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, op)
	}

	defaultPos := 0
	if highestPosition == nil {
		highestPosition = &defaultPos
	}

	query = "INSERT INTO Groups (GroupID, Name, Color, Position, Permissions) VALUES ($1, $2, $3, $4, $5);"

	if _, err := r.db.Exec(query, 1, newGroup.Name, newGroup.Color, *highestPosition+1, newGroup.Permissions); err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *groupRepo) fetch(ctx context.Context, query string, args ...interface{}) ([]*domain.Group, error) {
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

	results := make([]*domain.Group, 0)
	for rows.Next() {
		group := &domain.Group{}

		if err := r.scanRows(rows, group); err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.Wrap(domain.ErrNotFound, op)
			}

			return nil, errors.Wrap(err, op)
		}

		results = append(results, group)
	}

	return results, nil
}

// Store stores a new group in the database. The following fields must be set on the passed in group:
// Name, Color, Position, Permissions
func (r *groupRepo) Store(ctx context.Context, group *domain.Group) error {
	const op = opTag + "Store"

	query := "INSERT INTO Groups (Name, Color, Position, Permissions) VALUES ($1, $2, $3, $4);"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	res, err := stmt.ExecContext(ctx, group.Name, group.Color, group.Position, group.Permissions)
	if err != nil {
		r.logger.Error("Could not execute prepared statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	id, err := res.LastInsertId()
	if err != nil {
		r.logger.Error("Could not get ID of newly inserted community", zap.Error(err))
		return errors.Wrap(err, op)
	}

	group.ID = id

	return nil
}

func (r *groupRepo) GetAll(ctx context.Context) ([]*domain.Group, error) {
	const op = opTag + "GetAll"

	query := "SELECT * FROM Groups;"

	results, err := r.fetch(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return results, nil
}

func (r *groupRepo) GetByID(ctx context.Context, id int64) (*domain.Group, error) {
	const op = opTag + "GetByID"

	query := "SELECT * FROM Groups WHERE GroupID = $1;"

	results, err := r.fetch(ctx, query, id)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		return results[0], nil
	}

	return nil, errors.Wrap(domain.ErrNotFound, op)
}

func (r *groupRepo) GetUserGroups(ctx context.Context, userID string) ([]*domain.Group, error) {
	const op = opTag + "GetUserGroups"

	query := "SELECT * FROM UserGroups WHERE UserID = $1;"

	results, err := r.fetch(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if len(results) > 0 {
		return results, nil
	}

	return nil, domain.ErrNotFound
}

func (r *groupRepo) SetUserOverrides(ctx context.Context, userID string, overrides *domain.Overrides) error {
	const op = opTag + "SetUserOverrides"

	query := "UPDATE UserOverrides SET AllowOverrides = $1, DenyOverrides = $2 WHERE UserID = $3;"

	_, err := r.db.ExecContext(ctx, overrides.AllowOverrides, overrides.DenyOverrides, userID)
	if err != nil {
		r.logger.Error("Could not execute query", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	return nil
}

func (r *groupRepo) GetUserOverrides(ctx context.Context, userID string) (*domain.Overrides, error) {
	const op = opTag + "GetUserOverrides"

	query := "SELECT AllowOverrides, DenyOverrides FROM UserOverrides WHERE UserID = $1 LIMIT 1;"

	row := r.db.QueryRowContext(ctx, query, userID)

	overrides := &domain.Overrides{}

	if err := row.Scan(&overrides.AllowOverrides, &overrides.DenyOverrides); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(domain.ErrNotFound, op)
		}

		r.logger.Error("Could not scan user overrides", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	return overrides, nil
}

// Scan helpers
func (r *groupRepo) scanRow(row *sql.Row, group *domain.Group) error {
	return row.Scan(&group.ID, &group.Name, &group.Color, &group.Position, &group.Permissions, &group.CreatedAt, &group.ModifiedAt)
}

func (r *groupRepo) scanRows(rows *sql.Rows, group *domain.Group) error {
	return rows.Scan(&group.ID, &group.Name, &group.Color, &group.Position, &group.Permissions, &group.CreatedAt, &group.ModifiedAt)
}
