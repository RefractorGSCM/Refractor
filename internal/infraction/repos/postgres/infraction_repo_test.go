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
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/franela/goblin"
	"github.com/guregu/null"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"regexp"
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	var cols = []string{
		"InfractionID",
		"PlayerID",
		"Platform",
		"UserID",
		"ServerID",
		"Type",
		"Reason",
		"Duration",
		"SystemAction",
		"CreatedAt",
		"ModifiedAt",
	}
	var ctx = context.TODO()

	g.Describe("Postgres Infraction Repo", func() {
		var repo domain.InfractionRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo = NewInfractionRepo(db, zap.NewNop())
		})

		g.AfterEach(func() {
			_ = db.Close()
		})

		g.Describe("Store()", func() {
			g.Describe("Successful store", func() {
				var expected *domain.Infraction

				g.BeforeEach(func() {
					i := &domain.Infraction{
						InfractionID: 1,
						PlayerID:     "playerid",
						Platform:     "platform",
						UserID:       null.NewString("userid", true),
						ServerID:     2,
						Type:         domain.InfractionTypeKick,
						Reason:       null.NewString("Test reason", true),
						Duration:     null.NewInt(0, false),
						SystemAction: false,
						CreatedAt:    null.NewTime(time.Time{}, true),
						ModifiedAt:   null.NewTime(time.Time{}, false),
					}

					expected = i

					mock.ExpectPrepare("INSERT INTO Infractions")
					mock.ExpectQuery("INSERT INTO Infractions").WillReturnRows(
						sqlmock.NewRows(cols).
							AddRow(i.InfractionID, i.PlayerID, i.Platform, i.UserID, i.ServerID,
								i.Type, i.Reason, i.Duration, i.SystemAction, i.CreatedAt, i.ModifiedAt))
				})

				g.It("Should not return an error", func() {
					_, err := repo.Store(ctx, expected)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct infraction", func() {
					infr, err := repo.Store(ctx, expected)

					Expect(err).To(BeNil())
					Expect(infr).To(Equal(expected))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectPrepare("INSERT INTO Infractions")
					mock.ExpectQuery("INSERT INTO Infractions").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error and a nil infraction", func() {
					infr, err := repo.Store(ctx, &domain.Infraction{})

					Expect(err).ToNot(BeNil())
					Expect(infr).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetByID()", func() {
			g.Describe("Infraction found in db", func() {
				var mockInfraction *domain.Infraction
				var mockRows *sqlmock.Rows

				g.BeforeEach(func() {
					i := &domain.Infraction{
						InfractionID: 1,
						PlayerID:     "playerid",
						Platform:     "platform",
						UserID:       null.NewString("", false),
						ServerID:     1,
						Type:         domain.InfractionTypeMute,
						Reason:       null.NewString("Test reason", true),
						Duration:     null.NewInt(60, true),
						SystemAction: false,
						CreatedAt:    null.NewTime(time.Now(), true),
						ModifiedAt:   null.Time{},
					}

					mockInfraction = i

					mockRows = sqlmock.NewRows(cols).
						AddRow(i.InfractionID, i.PlayerID, i.Platform, i.UserID, i.ServerID, i.Type, i.Reason, i.Duration,
							i.SystemAction, i.CreatedAt, i.ModifiedAt)

					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Infractions")).WillReturnRows(mockRows)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetByID(ctx, mockInfraction.InfractionID)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct row scanned into an infraction", func() {
					found, err := repo.GetByID(ctx, mockInfraction.InfractionID)

					Expect(err).To(BeNil())
					Expect(found).To(Equal(mockInfraction))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Infraction not found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Infractions")).WillReturnRows(sqlmock.NewRows(cols))
				})

				g.It("Should return domain.ErrNotFound error", func() {
					_, err := repo.GetByID(ctx, 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Update()", func() {
			g.BeforeEach(func() {
				mock.ExpectPrepare("UPDATE Infractions SET")
			})

			g.Describe("Target row found", func() {
				var updatedInfraction *domain.Infraction
				var updateArgs domain.UpdateArgs

				g.BeforeEach(func() {
					ui := &domain.Infraction{
						InfractionID: 1,
						PlayerID:     "playerid",
						Platform:     "platform",
						UserID:       null.NewString("", false),
						ServerID:     1,
						Type:         domain.InfractionTypeBan,
						Reason:       null.NewString("Test reason", true),
						Duration:     null.NewInt(1440, true),
						SystemAction: false,
						CreatedAt:    null.NewTime(time.Now(), true),
						ModifiedAt:   null.Time{},
					}

					updatedInfraction = ui

					updateArgs = domain.UpdateArgs{
						"Reason":   null.NewString("Updated reason", true),
						"Duration": null.NewInt(120, true),
					}

					mock.ExpectQuery("UPDATE Infractions SET").WillReturnRows(sqlmock.NewRows(cols).
						AddRow(ui.InfractionID, ui.PlayerID, ui.Platform, ui.UserID, ui.ServerID, ui.Type, ui.Reason,
							ui.Duration, ui.SystemAction, ui.CreatedAt, ui.ModifiedAt))
				})

				g.It("Should not return an error", func() {
					_, err := repo.Update(ctx, updatedInfraction.InfractionID, updateArgs)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should scan and return the correct infraction", func() {
					updated, err := repo.Update(ctx, updatedInfraction.InfractionID, updateArgs)

					Expect(err).To(BeNil())
					Expect(updated).To(Equal(updatedInfraction))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Target row not found", func() {
				var updateArgs domain.UpdateArgs

				g.BeforeEach(func() {
					updateArgs = domain.UpdateArgs{
						"Reason": null.NewString("Updated reason", true),
					}

					mock.ExpectQuery("UPDATE Infractions SET").WillReturnError(sql.ErrNoRows)
				})

				g.It("Should return a domain.ErrNotFound error and a nil infraction", func() {
					infr, err := repo.Update(ctx, 1, updateArgs)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(infr).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Delete()", func() {
			g.Describe("Target infraction exists", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM Infractions").WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.Delete(ctx, 1)

					Expect(err).To(BeNil())
				})
			})

			g.Describe("Target infraction does not exist", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM Infractions").WillReturnResult(sqlmock.NewResult(0, 0))
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					err := repo.Delete(ctx, 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				})
			})
		})
	})
}
