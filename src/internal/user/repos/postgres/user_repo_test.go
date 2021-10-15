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

		g.Describe("IsDeactivated()", func() {
			g.Describe("User is deactivated", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM UserMeta")).WillReturnRows(
						sqlmock.NewRows([]string{"exists"}).AddRow(true))
				})

				g.It("Should not return an error", func() {
					_, err := repo.IsDeactivated(ctx, "userid")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return true", func() {
					isDeactivated, _ := repo.IsDeactivated(ctx, "userid")

					Expect(isDeactivated).To(BeTrue())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("User is not deactivated", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM UserMeta")).WillReturnRows(
						sqlmock.NewRows([]string{"exists"}).AddRow(false))
				})

				g.It("Should not return an error", func() {
					_, err := repo.IsDeactivated(ctx, "userid")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return false", func() {
					isDeactivated, _ := repo.IsDeactivated(ctx, "userid")

					Expect(isDeactivated).To(BeFalse())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetUsername()", func() {
			g.Describe("UserMeta is cached", func() {
				var cachedUser *domain.UserMeta

				g.BeforeEach(func() {
					cachedUser = &domain.UserMeta{
						ID:              "userid",
						InitialUsername: "initialUsername",
						Username:        "currentUsername",
						Deactivated:     false,
					}

					repo.cache.SetDefault(cachedUser.ID, cachedUser)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetUsername(ctx, cachedUser.ID)

					Expect(err).To(BeNil())
				})

				g.It("Should return username from cache without hitting DB", func() {
					username, err := repo.GetUsername(ctx, cachedUser.ID)

					Expect(err).To(BeNil())
					Expect(username).To(Equal(cachedUser.Username))
				})
			})

			g.Describe("UserMeta is not cached", func() {
				g.Describe("Username found in DB", func() {
					g.BeforeEach(func() {
						mock.ExpectQuery("SELECT Username FROM UserMeta").WillReturnRows(
							sqlmock.NewRows([]string{"username"}).AddRow("currentUsername"))
					})

					g.It("Should not return an error", func() {
						_, err := repo.GetUsername(ctx, "userid")

						Expect(err).To(BeNil())
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})

					g.It("Should return the correct username", func() {
						username, err := repo.GetUsername(ctx, "userid")

						Expect(err).To(BeNil())
						Expect(username).To(Equal("currentUsername"))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				g.Describe("Username not found in DB", func() {
					g.BeforeEach(func() {
						mock.ExpectQuery("SELECT Username FROM UserMeta").WillReturnError(sql.ErrNoRows)
					})

					g.It("Should return a domain.ErrNotFound error", func() {
						_, err := repo.GetUsername(ctx, "userid")

						Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})
			})
		})

		g.Describe("LinkPlayer()", func() {
			g.Describe("Successful link", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO UserPlayers").WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.LinkPlayer(ctx, "userid", "platform", "playerid")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO UserPlayers").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.LinkPlayer(ctx, "userid", "platform", "playerid")

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("UnlinkPlayer()", func() {
			g.Describe("Successful unlink", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM UserPlayers WHERE").WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.UnlinkPlayer(ctx, "userid", "platform", "playerid")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Link not found", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM UserPlayers WHERE").WillReturnResult(sqlmock.NewResult(0, 0))
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					err := repo.UnlinkPlayer(ctx, "userid", "platform", "playerid")

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM UserPlayers WHERE").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.UnlinkPlayer(ctx, "userid", "platform", "playerid")

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetLinkedPlayers()", func() {
			var expected []*domain.Player
			var playerCols []string

			g.BeforeEach(func() {
				expected = make([]*domain.Player, 0)
				playerCols = []string{"PlayerID", "Platform", "Watched", "LastSeen", "CreatedAt", "ModifiedAt"}
			})

			g.Describe("Linked players found", func() {
				g.BeforeEach(func() {
					expected = append(expected, &domain.Player{
						PlayerID:   "playerid",
						Platform:   "platform",
						Watched:    false,
						LastSeen:   time.Time{},
						CreatedAt:  time.Time{},
						ModifiedAt: time.Time{},
					}, &domain.Player{
						PlayerID:   "playerid2",
						Platform:   "platform2",
						Watched:    true,
						LastSeen:   time.Time{},
						CreatedAt:  time.Time{},
						ModifiedAt: time.Time{},
					})

					rows := sqlmock.NewRows(playerCols)

					for _, p := range expected {
						rows.AddRow(p.PlayerID, p.Platform, p.Watched, p.LastSeen, p.CreatedAt, p.ModifiedAt)
					}

					mock.ExpectQuery(regexp.QuoteMeta("SELECT p.* FROM UserPlayers up")).WillReturnRows(rows)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetLinkedPlayers(ctx, "userid")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct players", func() {
					got, err := repo.GetLinkedPlayers(ctx, "userid")

					Expect(err).To(BeNil())
					Expect(got).To(Equal(expected))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No linked players found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT p.* FROM UserPlayers up")).WillReturnRows(sqlmock.NewRows(playerCols))
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetLinkedPlayers(ctx, "userid")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return an empty player slice", func() {
					got, err := repo.GetLinkedPlayers(ctx, "userid")

					Expect(err).To(BeNil())
					Expect(got).To(Equal(expected))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT p.* FROM UserPlayers up")).WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := repo.GetLinkedPlayers(ctx, "userid")

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})
	})
}
