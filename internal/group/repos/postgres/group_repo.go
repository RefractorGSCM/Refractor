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
	"Refractor/pkg/querybuilders/psqlqb"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/gob"
	"fmt"
	"github.com/lib/pq"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math"
	"os"
	"time"
)

const opTag = "GroupRepo.Postgres."

type groupRepo struct {
	db     *sql.DB
	logger *zap.Logger
	cache  *gocache.Cache
	qb     domain.QueryBuilder
}

const cacheKeyBaseGroup = "base_group"

func NewGroupRepo(db *sql.DB, logger *zap.Logger) (domain.GroupRepo, error) {
	repo := &groupRepo{
		db:     db,
		logger: logger,
		cache:  gocache.New(30*time.Minute, 1*time.Hour),
		qb:     psqlqb.NewPostgresQueryBuilder(),
	}

	return repo, nil
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
		group := &domain.DBGroup{}

		if err := r.scanRows(rows, group); err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.Wrap(domain.ErrNotFound, op)
			}

			return nil, errors.Wrap(err, op)
		}

		results = append(results, group.Group())
	}

	return results, nil
}

// Store stores a new group in the database. The following fields must be set on the passed in group:
// Name, Color, Position, Permissions
func (r *groupRepo) Store(ctx context.Context, group *domain.Group) error {
	const op = opTag + "Store"

	query := "INSERT INTO Groups (Name, Color, Position, Permissions) VALUES ($1, $2, $3, $4) RETURNING GroupID;"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, group.Name, group.Color, group.Position, group.Permissions)

	var id int64

	if err := row.Scan(&id); err != nil {
		r.logger.Error("Could not execute prepared statement", zap.String("query", query), zap.Error(err))
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

	query := `SELECT * FROM Groups WHERE GroupID = $1;`

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

	query := `	SELECT
					g.*
				FROM UserGroups ug
				INNER JOIN Groups g ON g.GroupID = ug.GroupID
				WHERE UserID = $1 ORDER BY g.Position ASC;`

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

	query := `INSERT INTO UserOverrides (UserID, AllowOverrides, DenyOverrides) VALUES ($1, $2, $3)
				ON CONFLICT (UserID) DO UPDATE SET AllowOverrides = $4, DenyOverrides = $5;`

	_, err := r.db.ExecContext(ctx, query, userID, overrides.AllowOverrides, overrides.DenyOverrides,
		overrides.AllowOverrides, overrides.DenyOverrides)
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

var defaultDefaultGroup = &domain.Group{
	ID:          -1,
	Name:        "Everyone",
	Color:       0xb5b5b5,
	Position:    math.MaxInt32,
	Permissions: perms.GetDefaultPermissions().String(),
}

func (r *groupRepo) GetBaseGroup(ctx context.Context) (*domain.Group, error) {
	const op = opTag + "GetBaseGroup"

	// Check if base group exists in cache. If it does, return it and skip the IO.
	if bg, found := r.cache.Get(cacheKeyBaseGroup); found {
		baseGroup := bg.(*domain.Group)

		return baseGroup, nil
	}

	// Check if data file exists
	if _, err := os.Stat("./data/default_group.gob"); os.IsNotExist(err) {
		// If it doesn't, use SetBaseGroup to create it
		if err := r.SetBaseGroup(ctx, defaultDefaultGroup); err != nil {
			return nil, errors.Wrap(err, op)
		}
	}

	// Open data file and decode the data within
	file, err := os.Open("./data/default_group.gob")
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	defer func() {
		_ = file.Close()
	}()

	decoder := gob.NewDecoder(file)

	baseGroup := &domain.Group{}
	if err := decoder.Decode(baseGroup); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Set base group in cache
	r.cache.Set(cacheKeyBaseGroup, baseGroup, 1*time.Hour)

	return baseGroup, nil
}

func (r *groupRepo) SetBaseGroup(ctx context.Context, group *domain.Group) error {
	const op = opTag + "SetBaseGroup"

	// Check if data directory exists
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		if err := os.Mkdir("./data", os.ModePerm); err != nil {
			return errors.Wrap(err, op)
		}
	}

	// Create data file
	file, err := os.Create("./data/default_group.gob")
	if err != nil {
		return errors.Wrap(err, op)
	}

	defer func() {
		_ = file.Close()
	}()

	// Gob encode the group struct
	encoder := gob.NewEncoder(file)

	if err := encoder.Encode(group); err != nil {
		return errors.Wrap(err, op)
	}

	// Update cache
	r.cache.Set(cacheKeyBaseGroup, group, 1*time.Hour)

	return nil
}

func (r *groupRepo) Delete(ctx context.Context, id int64) error {
	const op = opTag + "Delete"

	query := "DELETE FROM Groups WHERE GroupID = $1;"

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

func (r *groupRepo) Update(ctx context.Context, id int64, args domain.UpdateArgs) (*domain.Group, error) {
	const op = opTag + "Update"

	query, values := r.qb.BuildUpdateQuery("Groups", id, "GroupID", args)

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, values...)

	updatedGroup := &domain.DBGroup{}
	if err := r.scanRow(row, updatedGroup); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(domain.ErrNotFound, op)
		}

		r.logger.Error("Could not scan updated group", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	return updatedGroup.Group(), nil
}

type GroupReorderInfo struct {
	*domain.GroupReorderInfo
}

func (gri *GroupReorderInfo) Value() (driver.Value, error) {
	return fmt.Sprintf("(%d, %d)", gri.GroupID, gri.NewPos), nil
}

func (r *groupRepo) Reorder(ctx context.Context, newPositions []*domain.GroupReorderInfo) error {
	const op = opTag + "Reorder"

	// Convert passed in group reordering info to our local GroupReorderInfo type
	var reorderInfo []*GroupReorderInfo
	for _, np := range newPositions {
		reorderInfo = append(reorderInfo, &GroupReorderInfo{
			GroupReorderInfo: &domain.GroupReorderInfo{
				GroupID: np.GroupID,
				NewPos:  np.NewPos,
			},
		})
	}

	query := `SELECT reorder_groups($1::reorder_groups_info[]);`

	_, err := r.db.QueryContext(ctx, query, pq.Array(reorderInfo))
	if err != nil {
		r.logger.Error("Could not execute reorder groups function", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	return nil
}

// Scan helpers
func (r *groupRepo) scanRow(row *sql.Row, group *domain.DBGroup) error {
	return row.Scan(&group.ID, &group.Name, &group.Color, &group.Position, &group.Permissions, &group.CreatedAt, &group.ModifiedAt)
}

func (r *groupRepo) scanRows(rows *sql.Rows, group *domain.DBGroup) error {
	return rows.Scan(&group.ID, &group.Name, &group.Color, &group.Position, &group.Permissions, &group.CreatedAt, &group.ModifiedAt)
}
