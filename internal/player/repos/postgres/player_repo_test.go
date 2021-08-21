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
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	var playerCols = []string{"PlayerID", "Platform", "Watched", "LastSeen", "CreatedAt", "ModifiedAt"}
	var playerNameCols = []string{"Name"}
	var ctx = context.TODO()

	g.Describe("Player Postgres Repo", func() {
		var repo domain.PlayerRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo = NewPlayerRepo(db, zap.NewNop())
		})

		g.After(func() {
			_ = db.Close()
		})

		g.Describe("Store()", func() {
			var player *domain.Player

			g.BeforeEach(func() {
				player = &domain.Player{
					PlayerID:    "playerid",
					Platform:    "platform",
					CurrentName: "testplayer",
				}

				mock.ExpectPrepare("INSERT INTO Players")
			})

			g.Describe("Success", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO Players").WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectPrepare("INSERT INTO PlayerNames")
					mock.ExpectExec("INSERT INTO PlayerNames").WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.Store(ctx, player)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Player insert error", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO Players").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.Store(ctx, player)

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("PlayerNames insert error", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO Players").WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectPrepare("INSERT INTO PlayerNames")
					mock.ExpectExec("INSERT INTO PlayerNames").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.Store(ctx, player)

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetByID()", func() {
			var player *domain.Player

			g.BeforeEach(func() {
				player = &domain.Player{
					PlayerID:   "playerid",
					Platform:   "platform",
					Watched:    true,
					LastSeen:   time.Now(),
					CreatedAt:  time.Now(),
					ModifiedAt: time.Now(),
				}
			})

			g.Describe("Player found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Players")).WillReturnRows(sqlmock.NewRows(playerCols).
						AddRow(player.PlayerID, player.Platform, player.Watched, player.LastSeen, player.CreatedAt, player.ModifiedAt))
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetByID(ctx, player.Platform, player.PlayerID)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct scanned player", func() {
					p, err := repo.GetByID(ctx, player.Platform, player.PlayerID)

					Expect(err).To(BeNil())
					Expect(p).To(Equal(player))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Player not found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Players")).WillReturnRows(sqlmock.NewRows(playerCols))
				})

				g.It("Should return domain.ErrNotFound error", func() {
					_, err := repo.GetByID(ctx, player.Platform, player.PlayerID)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return a nil player", func() {
					p, _ := repo.GetByID(ctx, player.Platform, player.PlayerID)

					Expect(p).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Exists()", func() {
			g.Describe("Player exists", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).WillReturnRows(sqlmock.
						NewRows([]string{"Exists"}).AddRow(true))
				})

				g.It("Should not return an error", func() {
					_, err := repo.Exists(ctx, domain.FindArgs{})

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return true", func() {
					exists, err := repo.Exists(ctx, domain.FindArgs{})

					Expect(err).To(BeNil())
					Expect(exists).To(BeTrue())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Player does not exist", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).WillReturnRows(sqlmock.
						NewRows([]string{"Exists"}).AddRow(false))
				})

				g.It("Should not return an error", func() {
					_, err := repo.Exists(ctx, domain.FindArgs{})

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return false", func() {
					exists, err := repo.Exists(ctx, domain.FindArgs{})

					Expect(err).To(BeNil())
					Expect(exists).To(BeFalse())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("UpdateName()", func() {
			var player *domain.Player

			g.BeforeEach(func() {
				player = &domain.Player{
					PlayerID: "playerid",
					Platform: "platform",
				}
			})

			g.Describe("Success", func() {
				var newName = "newName"

				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO PlayerNames").WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.UpdateName(ctx, player, newName)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should update the player struct to have the new name", func() {
					err := repo.UpdateName(ctx, player, newName)

					Expect(err).To(BeNil())
					Expect(player.CurrentName).To(Equal(newName))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Insertion error", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO PlayerNames").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.UpdateName(ctx, player, "newName")

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Update()", func() {
			g.BeforeEach(func() {
				mock.ExpectPrepare("UPDATE Players SET")
			})

			g.Describe("Target player found", func() {
				var updatedPlayer *domain.Player
				var updateArgs domain.UpdateArgs

				g.BeforeEach(func() {
					up := &domain.Player{
						PlayerID:  "id",
						Platform:  "platform",
						Watched:   true,
						LastSeen:  time.Now(),
						CreatedAt: time.Now(),
					}

					updateArgs = domain.UpdateArgs{
						"Watched": up.Watched,
					}

					updatedPlayer = up

					mock.ExpectQuery("UPDATE Players SET").WillReturnRows(sqlmock.NewRows(playerCols).
						AddRow(up.PlayerID, up.Platform, up.Watched, up.LastSeen, up.CreatedAt, up.ModifiedAt))
				})

				g.It("Should not return an error", func() {
					_, err := repo.Update(context.TODO(), updatedPlayer.Platform, updatedPlayer.PlayerID, updateArgs)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should scan and return the modified player", func() {
					updated, err := repo.Update(context.TODO(), updatedPlayer.Platform, updatedPlayer.PlayerID, updateArgs)

					Expect(err).To(BeNil())
					Expect(updated).To(Equal(updatedPlayer))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Target player not found", func() {
				var updateArgs domain.UpdateArgs

				g.BeforeEach(func() {
					updateArgs = domain.UpdateArgs{
						"Watched": false,
					}

					mock.ExpectQuery("UPDATE Players SET").WillReturnError(sql.ErrNoRows)
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					_, err := repo.Update(context.TODO(), "platform", "playerid", updateArgs)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return a nil player", func() {
					p, err := repo.Update(context.TODO(), "platform", "playerid", updateArgs)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(p).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetPlayerNames()", func() {
			g.Describe("Player names found", func() {
				var expected []string

				g.BeforeEach(func() {
					expected = []string{
						"name1",
						"name2",
						"name3",
						"name4",
					}

					mock.ExpectQuery("SELECT Name FROM PlayerNames WHERE").WillReturnRows(sqlmock.
						NewRows(playerNameCols).
						AddRow(expected[0]).
						AddRow(expected[1]).
						AddRow(expected[2]).
						AddRow(expected[3]))
				})

				g.It("Should not return an error", func() {
					_, _, err := repo.GetPlayerNames(ctx, "id", "platform")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct names", func() {
					currentName, previousNames, err := repo.GetPlayerNames(ctx, "id", "platform")

					Expect(err).To(BeNil())
					Expect(currentName).To(Equal(expected[0]))
					Expect(previousNames).To(Equal(expected[1:]))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Only one name was found", func() {
				var expected []string

				g.BeforeEach(func() {
					expected = []string{
						"name1",
					}

					mock.ExpectQuery("SELECT Name FROM PlayerNames WHERE").WillReturnRows(sqlmock.
						NewRows(playerNameCols).
						AddRow(expected[0]))
				})

				g.It("Should not return an error", func() {
					_, _, err := repo.GetPlayerNames(ctx, "id", "platform")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct names", func() {
					currentName, previousNames, err := repo.GetPlayerNames(ctx, "id", "platform")

					Expect(err).To(BeNil())
					Expect(currentName).To(Equal(expected[0]))
					Expect(previousNames).To(Equal(expected[1:]))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})
	})
}
