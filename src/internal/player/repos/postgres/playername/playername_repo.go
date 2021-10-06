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

package playername

import (
	"Refractor/domain"
	"Refractor/pkg/querybuilders/psqlqb"
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

const opTag = "PlayerRepo.Postgres."

type playerNameRepo struct {
	db     *sql.DB
	logger *zap.Logger
	qb     domain.QueryBuilder
}

func NewPlayerNameRepo(db *sql.DB, logger *zap.Logger) domain.PlayerNameRepo {
	return &playerNameRepo{
		db:     db,
		logger: logger,
		qb:     psqlqb.NewPostgresQueryBuilder(),
	}
}

func (r *playerNameRepo) Store(ctx context.Context, id, platform, name string) error {
	const op = opTag + "Store"

	// Insert into PlayerNames
	query := `INSERT INTO PlayerNames (PlayerID, Platform, Name, DateRecorded) VALUES ($1, $2, $3, $4)
			ON CONFLICT (PlayerID, Platform, Name) DO UPDATE SET DateRecorded = CURRENT_TIMESTAMP;`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("Could not prepare insert statement", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	if _, err := stmt.ExecContext(ctx, id, platform, name, time.Now()); err != nil {
		r.logger.Error("Could not execute insert query", zap.String("query", query), zap.Error(err))
		return errors.Wrap(err, op)
	}

	return nil
}

// UpdateName updates the player's name in the database. It does this by adding a new row in PlayerNames
// if the player never used this name before. If they have used this name before, it updates the DateRecorded
// field to the current timestamp.
//
// The following fields are required on the passed in player: PlayerID, Platform
//
// The passed in player struct has it's CurrentName field updated on success.
func (r *playerNameRepo) UpdateName(ctx context.Context, player *domain.Player, newName string) error {
	const op = opTag + "UpdateName"

	query := `INSERT INTO PlayerNames (PlayerID, Platform, Name, DateRecorded) VALUES ($1, $2, $3, $4)
			ON CONFLICT (PlayerID, Platform, Name) DO UPDATE SET DateRecorded = CURRENT_TIMESTAMP;`

	runeName := []rune(newName)

	if _, err := r.db.ExecContext(ctx, query, player.PlayerID, player.Platform, string(runeName), time.Now()); err != nil {
		r.logger.Error("Could not execute name update query", zap.Error(err))
		return errors.Wrap(err, op)
	}

	player.CurrentName = newName

	return nil
}

func (r *playerNameRepo) GetNames(ctx context.Context, id, platform string) (string, []string, error) {
	const op = opTag + "GetNames"

	query := "SELECT Name FROM PlayerNames WHERE PlayerID = $1 AND Platform = $2 ORDER BY DateRecorded DESC;"

	rows, err := r.db.QueryContext(ctx, query, id, platform)
	if err != nil {
		return "", nil, err
	}

	var names []string

	for rows.Next() {
		name := ""

		err = rows.Scan(&name)
		if err != nil {
			r.logger.Error("Could not scan player name", zap.Error(err))
			return "", nil, errors.Wrap(err, op)
		}

		names = append(names, name)
	}

	if names == nil {
		return "", nil, errors.Wrap(domain.ErrNotFound, op)
	}

	// PlayerNames are ordered in descending order by DateRecorded so index 0 will be the most recent name
	return names[0], names[1:], nil
}
