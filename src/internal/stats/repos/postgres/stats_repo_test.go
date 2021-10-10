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
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"regexp"
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Postgres Stats Repo", func() {
		var repo domain.StatsRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB
		var ctx context.Context

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo = NewStatsRepo(db, zap.NewNop())
			ctx = context.TODO()
		})

		g.After(func() {
			_ = db.Close()
		})

		g.Describe("GetTotalPlayers()", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(1) FROM Players")).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(60))
			})

			g.It("Should not return an error", func() {
				_, err := repo.GetTotalPlayers(ctx)

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return the correct count", func() {
				count, err := repo.GetTotalPlayers(ctx)

				Expect(err).To(BeNil())
				Expect(count).To(Equal(60))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("GetTotalInfractions()", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(1) FROM Infractions")).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(60))
			})

			g.It("Should not return an error", func() {
				_, err := repo.GetTotalInfractions(ctx)

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return the correct count", func() {
				count, err := repo.GetTotalInfractions(ctx)

				Expect(err).To(BeNil())
				Expect(count).To(Equal(60))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("GetTotalNewPlayersInRange()", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(1) FROM Players")).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(60))
			})

			g.It("Should not return an error", func() {
				_, err := repo.GetTotalNewPlayersInRange(ctx, time.Now(), time.Now())

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return the correct count", func() {
				count, err := repo.GetTotalNewPlayersInRange(ctx, time.Now(), time.Now())

				Expect(err).To(BeNil())
				Expect(count).To(Equal(60))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("GetTotalOnlinePlayersInRange()", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(1) FROM Players")).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(60))
			})

			g.It("Should not return an error", func() {
				_, err := repo.GetTotalOnlinePlayersInRange(ctx, time.Now(), time.Now())

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return the correct count", func() {
				count, err := repo.GetTotalOnlinePlayersInRange(ctx, time.Now(), time.Now())

				Expect(err).To(BeNil())
				Expect(count).To(Equal(60))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})
	})
}
