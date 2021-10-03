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
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"regexp"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	var cols = []string{"WordID", "Word"}
	var ctx = context.TODO()

	g.Describe("Postgres Flagged Words Repo", func() {
		var repo domain.FlaggedWordRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo = NewFlaggedWordRepo(db, zap.NewNop())
		})

		g.After(func() {
			_ = db.Close()
		})

		g.Describe("Store()", func() {
			g.Describe("Successful store", func() {
				var flaggedWord *domain.FlaggedWord

				g.BeforeEach(func() {
					flaggedWord = &domain.FlaggedWord{
						Word: "word",
					}

					mock.ExpectPrepare("INSERT INTO FlaggedWords")
					mock.ExpectQuery("INSERT INTO FlaggedWords").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				})

				g.It("Should not return an error", func() {
					err := repo.Store(ctx, flaggedWord)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should set the correct ID on the flagged word", func() {
					err := repo.Store(ctx, flaggedWord)

					Expect(err).To(BeNil())
					Expect(flaggedWord.ID).To(Equal(int64(1)))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectPrepare("INSERT INTO FlaggedWords")
					mock.ExpectQuery("INSERT INTO FlaggedWords").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.Store(ctx, &domain.FlaggedWord{})

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetAll()", func() {
			g.Describe("Results found", func() {
				var expected []*domain.FlaggedWord

				g.BeforeEach(func() {
					expected = []*domain.FlaggedWord{
						{
							ID:   1,
							Word: "word1",
						},
						{
							ID:   2,
							Word: "word2",
						},
						{
							ID:   3,
							Word: "word3",
						},
						{
							ID:   4,
							Word: "word4",
						},
						{
							ID:   5,
							Word: "word5",
						},
						{
							ID:   6,
							Word: "word6",
						},
					}

					rows := sqlmock.NewRows(cols)
					for _, fw := range expected {
						rows.AddRow(fw.ID, fw.Word)
					}

					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM FlaggedWords")).WillReturnRows(rows)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetAll(ctx)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct scanned results", func() {
					got, err := repo.GetAll(ctx)

					Expect(err).To(BeNil())
					Expect(got).To(Equal(expected))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No results found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM FlaggedWords")).WillReturnRows(sqlmock.NewRows(cols))
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					_, err := repo.GetAll(ctx)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM FlaggedWords")).WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := repo.GetAll(ctx)

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Update()", func() {
			g.BeforeEach(func() {
				mock.ExpectPrepare("UPDATE FlaggedWords SET")
			})

			g.Describe("Successful update", func() {
				var updated *domain.FlaggedWord

				g.BeforeEach(func() {
					updated = &domain.FlaggedWord{
						ID:   2,
						Word: "updated word",
					}

					mock.ExpectQuery("UPDATE FlaggedWords SET").WillReturnRows(sqlmock.NewRows(cols).
						AddRow(updated.ID, updated.Word))
				})

				g.It("Should not return an error", func() {
					_, err := repo.Update(ctx, updated.ID, updated.Word)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should scan and return the correct infraction", func() {
					got, err := repo.Update(ctx, updated.ID, updated.Word)

					Expect(err).To(BeNil())
					Expect(got).To(Equal(updated))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Target row not found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery("UPDATE FlaggedWords SET").WillReturnRows(sqlmock.NewRows(cols))
				})

				g.It("Should return a doamin.ErrNotFound error and a nil FlaggedWord", func() {
					got, err := repo.Update(ctx, 1, "")

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(got).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Delete()", func() {
			g.Describe("Successful delete", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM FlaggedWords").WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.Delete(ctx, 1)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Target row not found", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM FlaggedWords").WillReturnResult(sqlmock.NewResult(0, 0))
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					err := repo.Delete(ctx, 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})
	})
}
