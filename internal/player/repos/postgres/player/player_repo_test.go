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

package player

import (
	"Refractor/domain"
	"Refractor/domain/mocks"
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
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
	var ctx = context.TODO()

	g.Describe("Player Postgres Repo", func() {
		var repo domain.PlayerRepo
		var mockRepo sqlmock.Sqlmock
		var nameRepo *mocks.PlayerNameRepo
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mockRepo, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			nameRepo = new(mocks.PlayerNameRepo)
			repo = NewPlayerRepo(db, nameRepo, zap.NewNop())
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

				mockRepo.ExpectPrepare("INSERT INTO Players")
			})

			g.Describe("Success", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectExec("INSERT INTO Players").WillReturnResult(sqlmock.NewResult(0, 1))
					nameRepo.On("Store", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				})

				g.It("Should not return an error", func() {
					err := repo.Store(ctx, player)

					Expect(err).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Player insert error", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectExec("INSERT INTO Players").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.Store(ctx, player)

					Expect(err).ToNot(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("PlayerNames store error", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectExec("INSERT INTO Players").WillReturnResult(sqlmock.NewResult(0, 1))
					nameRepo.On("Store", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.Store(ctx, player)

					Expect(err).ToNot(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetByID()", func() {
			var player *domain.Player

			g.BeforeEach(func() {
				player = &domain.Player{
					PlayerID:      "playerid",
					Platform:      "platform",
					Watched:       true,
					CurrentName:   "currentName",
					PreviousNames: []string{"prev1", "prev2"},
					LastSeen:      time.Now(),
					CreatedAt:     time.Now(),
					ModifiedAt:    time.Now(),
				}
			})

			g.Describe("Player found", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Players")).WillReturnRows(sqlmock.NewRows(playerCols).
						AddRow(player.PlayerID, player.Platform, player.Watched, player.LastSeen, player.CreatedAt, player.ModifiedAt))
					nameRepo.On("GetNames", mock.Anything, player.PlayerID, player.Platform).
						Return(player.CurrentName, player.PreviousNames, nil)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetByID(ctx, player.Platform, player.PlayerID)

					Expect(err).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct scanned player", func() {
					p, err := repo.GetByID(ctx, player.Platform, player.PlayerID)

					Expect(err).To(BeNil())
					Expect(p).To(Equal(player))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Returned player should have CurrentName and PreviousNames set", func() {

				})
			})

			g.Describe("Player not found", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Players")).WillReturnRows(sqlmock.NewRows(playerCols))
				})

				g.It("Should return domain.ErrNotFound error", func() {
					_, err := repo.GetByID(ctx, player.Platform, player.PlayerID)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return a nil player", func() {
					p, _ := repo.GetByID(ctx, player.Platform, player.PlayerID)

					Expect(p).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Exists()", func() {
			g.Describe("Player exists", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).WillReturnRows(sqlmock.
						NewRows([]string{"Exists"}).AddRow(true))
				})

				g.It("Should not return an error", func() {
					_, err := repo.Exists(ctx, domain.FindArgs{})

					Expect(err).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return true", func() {
					exists, err := repo.Exists(ctx, domain.FindArgs{})

					Expect(err).To(BeNil())
					Expect(exists).To(BeTrue())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Player does not exist", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).WillReturnRows(sqlmock.
						NewRows([]string{"Exists"}).AddRow(false))
				})

				g.It("Should not return an error", func() {
					_, err := repo.Exists(ctx, domain.FindArgs{})

					Expect(err).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return false", func() {
					exists, err := repo.Exists(ctx, domain.FindArgs{})

					Expect(err).To(BeNil())
					Expect(exists).To(BeFalse())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Update()", func() {
			g.BeforeEach(func() {
				mockRepo.ExpectPrepare("UPDATE Players SET")
			})

			g.Describe("Target player found", func() {
				var updatedPlayer *domain.Player
				var updateArgs domain.UpdateArgs

				g.BeforeEach(func() {
					up := &domain.Player{
						PlayerID:      "id",
						Platform:      "platform",
						CurrentName:   "currentName",
						PreviousNames: []string{"prev1", "prev2"},
						Watched:       true,
						LastSeen:      time.Now(),
						CreatedAt:     time.Now(),
					}

					updateArgs = domain.UpdateArgs{
						"Watched": up.Watched,
					}

					updatedPlayer = up

					mockRepo.ExpectQuery("UPDATE Players SET").WillReturnRows(sqlmock.NewRows(playerCols).
						AddRow(up.PlayerID, up.Platform, up.Watched, up.LastSeen, up.CreatedAt, up.ModifiedAt))
					nameRepo.On("GetNames", mock.Anything, up.PlayerID, up.Platform).
						Return(up.CurrentName, up.PreviousNames, nil)
				})

				g.It("Should not return an error", func() {
					_, err := repo.Update(context.TODO(), updatedPlayer.Platform, updatedPlayer.PlayerID, updateArgs)

					Expect(err).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should scan and return the modified player", func() {
					updated, err := repo.Update(context.TODO(), updatedPlayer.Platform, updatedPlayer.PlayerID, updateArgs)

					Expect(err).To(BeNil())
					Expect(updated).To(Equal(updatedPlayer))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Target player not found", func() {
				var updateArgs domain.UpdateArgs

				g.BeforeEach(func() {
					updateArgs = domain.UpdateArgs{
						"Watched": false,
					}

					mockRepo.ExpectQuery("UPDATE Players SET").WillReturnError(sql.ErrNoRows)
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					_, err := repo.Update(context.TODO(), "platform", "playerid", updateArgs)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return a nil player", func() {
					p, err := repo.Update(context.TODO(), "platform", "playerid", updateArgs)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(p).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})
		})
	})
}
