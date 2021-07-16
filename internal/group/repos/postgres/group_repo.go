package postgres

import (
	"Refractor/domain"
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

func NewGroupRepo(db *sql.DB, logger *zap.Logger) domain.GroupRepo {
	return &groupRepo{
		db:     db,
		logger: logger,
	}
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

	return nil, nil
}

// Scan helpers
func (r *groupRepo) scanRow(row *sql.Row, group *domain.Group) error {
	return row.Scan(&group.ID, &group.Name, &group.Color, &group.Position, &group.Permissions, &group.CreatedAt, &group.ModifiedAt)
}

func (r *groupRepo) scanRows(rows *sql.Rows, group *domain.Group) error {
	return rows.Scan(&group.ID, &group.Name, &group.Color, &group.Position, &group.Permissions, &group.CreatedAt, &group.ModifiedAt)
}
