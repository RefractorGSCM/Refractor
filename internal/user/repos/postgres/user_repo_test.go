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
	"Refractor/pkg/querybuilders/psqlqb"
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	gocache "github.com/patrickmn/go-cache"
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

	ctx := context.TODO()
	cols := []string{"UserID", "InitialUsername", "Username", "Deactivated"}

	g.Describe("User Repo", func() {
		var repo *userRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo = &userRepo{
				db:     db,
				logger: zap.NewNop(),
				qb:     psqlqb.NewPostgresQueryBuilder(),
				cache:  gocache.New(time.Minute*1, time.Minute*1),
			}
		})

		g.After(func() {
			_ = db.Close()
		})

		g.Describe("Store()", func() {
			g.BeforeEach(func() {
				mock.ExpectPrepare("INSERT INTO UserMeta")
			})

			g.Describe("Successful store", func() {
				var meta *domain.UserMeta

				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO UserMeta").WillReturnResult(sqlmock.NewResult(0, 1))

					meta = &domain.UserMeta{
						ID:              "userid",
						InitialUsername: "initial",
						Username:        "initial",
						Deactivated:     false,
					}
				})

				g.It("Should not return an error", func() {
					err := repo.Store(ctx, meta)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Storage error", func() {
				g.It("Should return an error on SQL error", func() {
					mock.ExpectExec(regexp.QuoteMeta("INSERT INTO UserMeta")).WillReturnError(fmt.Errorf("err"))

					meta := &domain.UserMeta{}
					err := repo.Store(ctx, meta)

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetByID()", func() {
			g.Describe("The user's metadata is not cached", func() {
				g.Describe("A result was found", func() {
					var meta *domain.UserMeta
					var rows *sqlmock.Rows

					g.BeforeEach(func() {
						meta = &domain.UserMeta{
							ID:              "userid",
							InitialUsername: "initial",
							Username:        "initial",
							Deactivated:     false,
						}

						rows = sqlmock.NewRows(cols).
							AddRow(meta.ID, meta.InitialUsername, meta.Username, meta.Deactivated)

						mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM UserMeta")).WillReturnRows(rows)
					})

					g.It("Should not return an error", func() {
						_, err := repo.GetByID(ctx, meta.ID)

						Expect(err).To(BeNil())
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})

					g.It("Should return the correct row scanned into a UserMeta struct", func() {
						scanned, _ := repo.GetByID(ctx, meta.ID)

						Expect(scanned).To(Equal(meta))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				g.Describe("No result was found", func() {
					g.BeforeEach(func() {
						mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM UserMeta")).WillReturnRows(sqlmock.NewRows(cols))
					})

					g.It("Should return domain.ErrNotFound error", func() {
						_, err := repo.GetByID(ctx, "notfound-userid")

						Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					})
				})
			})

			g.Describe("The user's metadata is cached", func() {
				g.It("Should return the cached user without interacting with the database", func() {
					meta := &domain.UserMeta{
						ID:              "userid",
						InitialUsername: "cached",
						Username:        "cached",
						Deactivated:     false,
					}

					repo.cache.SetDefault("userid", meta)

					cached, err := repo.GetByID(ctx, "userid")

					Expect(err).To(BeNil())
					Expect(cached).To(Equal(meta))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Update()", func() {
			g.BeforeEach(func() {
				mock.ExpectPrepare("UPDATE UserMeta SET")
			})

			g.Describe("Target row found", func() {
				var updatedMeta *domain.UserMeta
				var updateArgs domain.UpdateArgs

				g.BeforeEach(func() {
					um := &domain.UserMeta{
						ID:              "userid",
						InitialUsername: "initial",
						Username:        "initial",
						Deactivated:     false,
					}

					updateArgs = domain.UpdateArgs{
						"Username": "newusername",
					}

					updatedMeta = um

					mock.ExpectQuery("UPDATE UserMeta SET").WillReturnRows(sqlmock.NewRows(cols).
						AddRow(um.ID, um.InitialUsername, um.Username, um.Deactivated))
				})

				g.It("Should not return an error", func() {
					_, err := repo.Update(context.TODO(), updatedMeta.ID, updateArgs)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should scan and return the correct UserMeta", func() {
					updated, err := repo.Update(context.TODO(), updatedMeta.ID, updateArgs)

					Expect(err).To(BeNil())
					Expect(updated).To(Equal(updatedMeta))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should update the user's cached metadata", func() {
					updated, err := repo.Update(context.TODO(), updatedMeta.ID, updateArgs)

					Expect(err).To(BeNil())
					Expect(updated).To(Equal(updatedMeta))
					Expect(mock.ExpectationsWereMet()).To(BeNil())

					cached, _ := repo.cache.Get(updatedMeta.ID)
					Expect(cached).To(Equal(updated))
				})
			})

			g.Describe("Target row not found", func() {
				var updateArgs domain.UpdateArgs

				g.BeforeEach(func() {
					updateArgs = domain.UpdateArgs{
						"Username": "newusername",
					}

					mock.ExpectQuery("UPDATE UserMeta SET").WillReturnError(sql.ErrNoRows)
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					_, err := repo.Update(context.TODO(), "userid", updateArgs)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return a nil UserMeta", func() {
					g, _ := repo.Update(context.TODO(), "userid", updateArgs)

					Expect(g).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})
	})
}
