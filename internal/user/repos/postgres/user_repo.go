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
	"context"
	"database/sql"
	"go.uber.org/zap"
)

type userRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewUserRepo(db *sql.DB, log *zap.Logger) domain.UserRepo {
	return &userRepo{
		db:     db,
		logger: log,
	}
}

func (r *userRepo) Store(ctx context.Context, userInfo *domain.UserInfo) error {
	panic("implement me")
}

func (r *userRepo) GetByID(ctx context.Context, userID string) (*domain.UserInfo, error) {
	panic("implement me")
}

func (r *userRepo) SetUsername(ctx context.Context, username string) error {
	panic("implement me")
}
