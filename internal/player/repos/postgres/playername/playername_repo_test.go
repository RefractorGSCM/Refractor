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
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	var playerNameCols = []string{"Name"}
	var ctx = context.TODO()

	g.Describe("PlayerName Postgres Repo", func() {
		var repo domain.PlayerNameRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo = NewPlayerNameRepo(db, zap.NewNop())
		})

		g.After(func() {
			_ = db.Close()
		})

		g.Describe("Store()", func() {
			g.BeforeEach(func() {
				mock.ExpectPrepare("INSERT INTO PlayerNames")
			})

			g.Describe("Success", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO PlayerNames").WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.Store(ctx, "id", "platform", "name")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("PlayerName insert error", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO PlayerNames").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.Store(ctx, "id", "platform", "name")

					Expect(err).ToNot(BeNil())
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
					_, _, err := repo.GetNames(ctx, "id", "platform")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct names", func() {
					currentName, previousNames, err := repo.GetNames(ctx, "id", "platform")

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
					_, _, err := repo.GetNames(ctx, "id", "platform")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct names", func() {
					currentName, previousNames, err := repo.GetNames(ctx, "id", "platform")

					Expect(err).To(BeNil())
					Expect(currentName).To(Equal(expected[0]))
					Expect(previousNames).To(Equal(expected[1:]))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})
	})
}
