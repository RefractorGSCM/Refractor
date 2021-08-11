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
	"Refractor/pkg/conf"
	"Refractor/pkg/querybuilders/psqlqb"
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const opTag = "ServerRepo.Postgres."

type serverRepo struct {
	db     *sql.DB
	logger *zap.Logger
	qb     domain.QueryBuilder
	conf   *conf.Config
}

func NewServerRepo(db *sql.DB, logger *zap.Logger, conf *conf.Config) domain.ServerRepo {
	return &serverRepo{
		db:     db,
		logger: logger,
		qb:     psqlqb.NewPostgresQueryBuilder(),
		conf:   conf,
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
		server := &domain.DBServer{}

		if err := r.scanRows(rows, server); err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.Wrap(domain.ErrNotFound, op)
			}

			return nil, errors.Wrap(err, op)
		}

		results = append(results, server.Server())
	}

	return results, nil
}

// Store stores a new server in the database. The following fields must be set on the passed in server:
// Game, Name, Address, RCONPort, RCONPassword.
func (r *serverRepo) Store(ctx context.Context, server *domain.Server) error {
	const op = opTag + "Store"

	query := "INSERT INTO Servers (Game, Name, Address, RCONPort, RCONPassword) VALUES ($1, $2, $3, $4, $5) RETURNING ServerID;"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	// Encrypt the server's RCON password
	encrypted, err := aeshelper.Encrypt([]byte(server.RCONPassword), r.conf.EncryptionKey)
	if err != nil {
		r.logger.Error("Could not encrypt server RCON password", zap.Error(err))
		return errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, server.Game, server.Name, server.Address, server.RCONPort, encrypted)

	var id int64
	if err := row.Scan(&id); err != nil {
		r.logger.Error("Could not scan ServerID from row", zap.Error(err))
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

func (r *serverRepo) GetAll(ctx context.Context) ([]*domain.Server, error) {
	const op = opTag + "GetAll"

	query := "SELECT * FROM Servers;"

	results, err := r.fetch(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return results, nil
}

func (r *serverRepo) Deactivate(ctx context.Context, id int64) error {
	const op = opTag + "Deactivate"

	query := "UPDATE Servers SET Deactivated = TRUE WHERE ServerID = $1;"

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Could not execute query", zap.String("Query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Could not get RowsAffected", zap.Error(err))
		return errors.Wrap(err, op)
	}

	if affected < 1 {
		return errors.Wrap(domain.ErrNotFound, op)
	}

	return nil
}

func (r *serverRepo) Update(ctx context.Context, id int64, args domain.UpdateArgs) (*domain.Server, error) {
	const op = opTag + "Update"

	// If the RCON password is being updated, encrypt it.
	if args["RCONPassword"] != nil {
		encrypted, err := aeshelper.Encrypt([]byte(*args["RCONPassword"].(*string)), r.conf.EncryptionKey)
		if err != nil {
			r.logger.Error("Could not encrypt server RCON password", zap.Error(err))
			return nil, errors.Wrap(err, op)
		}

		args["RCONPassword"] = encrypted
	}

	query, values := r.qb.BuildUpdateQuery("Servers", id, "ServerID", args)

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare statement", zap.String("query", query), zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	row := stmt.QueryRowContext(ctx, values...)

	updatedServer := &domain.DBServer{}
	if err := r.scanRow(row, updatedServer); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(domain.ErrNotFound, op)
		}

		r.logger.Error("Could not scan updated server", zap.Error(err))
		return nil, errors.Wrap(err, op)
	}

	return updatedServer.Server(), nil
}

// Scan helpers
func (r *serverRepo) scanRow(row *sql.Row, server *domain.DBServer) error {
	err := row.Scan(&server.ID, &server.Game, &server.Name, &server.Address, &server.RCONPort, &server.RCONPassword, &server.Deactivated, &server.CreatedAt, &server.ModifiedAt)
	if err != nil {
		return err
	}

	// Decrypt the server's RCON password
	decrypted, err := aeshelper.Decrypt([]byte(server.RCONPassword), r.conf.EncryptionKey)
	if err != nil {
		r.logger.Error("Could not decrypt server RCON password", zap.Error(err))
		return err
	}

	server.RCONPassword = string(decrypted)
	return nil
}

func (r *serverRepo) scanRows(rows *sql.Rows, server *domain.DBServer) error {
	err := rows.Scan(&server.ID, &server.Game, &server.Name, &server.Address, &server.RCONPort, &server.RCONPassword, &server.Deactivated, &server.CreatedAt, &server.ModifiedAt)
	if err != nil {
		return err
	}

	// Decrypt the server's RCON password
	decrypted, err := aeshelper.Decrypt([]byte(server.RCONPassword), r.conf.EncryptionKey)
	if err != nil {
		r.logger.Error("Could not decrypt server RCON password", zap.Error(err))
		return err
	}

	server.RCONPassword = string(decrypted)
	return nil
}
