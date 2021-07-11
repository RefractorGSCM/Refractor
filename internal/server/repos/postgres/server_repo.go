package postgres

import (
	"Refractor/domain"
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const opTag = "ServerRepo.Postgres."

type serverRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewServerRepo(db *sql.DB, logger *zap.Logger) domain.ServerRepo {
	return &serverRepo{
		db:     db,
		logger: logger,
	}
}

func (r *serverRepo) fetch(ctx context.Context, query string, args ...interface{}) ([]*domain.Server, error) {
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

	results := make([]*domain.Server, 0)
	for rows.Next() {
		server := &domain.Server{}

		if err := r.scanRows(rows, server); err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.Wrap(domain.ErrNotFound, op)
			}

			return nil, errors.Wrap(err, op)
		}

		results = append(results, server)
	}

	return results, nil
}

// Store stores a new server in the database. The following fields must be set on the passed in server:
// Game, Name, Address, RCONPort, RCONPassword.
func (r *serverRepo) Store(ctx context.Context, server *domain.Server) error {
	const op = opTag + "Store"

	query := "INSERT INTO Servers (Game, Name, Address, RCONPort, RCONPassword) VALUES ($1, $2, $3, $4, $5);"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	res, err := stmt.ExecContext(ctx, server.Game, server.Name, server.Address, server.RCONPort, server.RCONPassword)
	if err != nil {
		r.logger.Error("Could not execute prepared statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	id, err := res.LastInsertId()
	if err != nil {
		r.logger.Error("Could not get ID of newly inserted community", zap.Error(err))
		return errors.Wrap(err, op)
	}

	server.ID = id

	return nil
}

func (r *serverRepo) GetByID(ctx context.Context, id int64) (*domain.Server, error) {
	const op = opTag + "GetByID"

	query := "SELECT * FROM Servers WHERE ServerID = $1;"

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
func (r *serverRepo) scanRow(row *sql.Row, server *domain.Server) error {
	return row.Scan(&server.ID, &server.Game, &server.Name, &server.Address, &server.RCONPort, &server.RCONPassword, &server.CreatedAt, &server.ModifiedAt)
}

func (r *serverRepo) scanRows(rows *sql.Rows, server *domain.Server) error {
	return rows.Scan(&server.ID, &server.Game, &server.Name, &server.Address, &server.RCONPort, &server.RCONPassword, &server.CreatedAt, &server.ModifiedAt)
}
