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
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"regexp"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	var cols = []string{"MessageID", "PlayerID", "Platform", "ServerID", "Message", "Flagged", "CreatedAt", "ModifiedAt"}

	g.Describe("ChatMessage Postgres Repo", func() {
		var repo *chatRepo
		var mockRepo sqlmock.Sqlmock
		var db *sql.DB
		var ctx context.Context

		g.BeforeEach(func() {
			var err error

			db, mockRepo, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo = &chatRepo{
				db:     db,
				logger: zap.NewNop(),
				qb:     psqlqb.NewPostgresQueryBuilder(),
			}

			ctx = context.TODO()
		})

		g.After(func() {
			_ = db.Close()
		})

		g.Describe("Store()", func() {
			var chatMessage *domain.ChatMessage

			g.BeforeEach(func() {
				chatMessage = &domain.ChatMessage{
					PlayerID: "playerid",
					Platform: "platform",
					ServerID: 1,
					Message:  "test chat message",
					Flagged:  true,
				}

				mockRepo.ExpectBegin()
				mockRepo.ExpectPrepare("INSERT INTO ChatMessages")
			})

			g.Describe("Successful store", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery("INSERT INTO ChatMessages").WillReturnRows(sqlmock.NewRows(
						[]string{"id"}).AddRow(1))
					mockRepo.ExpectCommit()
				})

				g.It("Should not return an error", func() {
					err := repo.Store(ctx, chatMessage)

					Expect(err).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should update the ID in the provided struct", func() {
					err := repo.Store(ctx, chatMessage)

					Expect(err).To(BeNil())
					Expect(chatMessage.MessageID).To(Equal(int64(1)))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Insert error", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery("INSERT INTO ChatMessages").WillReturnError(fmt.Errorf("err"))
					mockRepo.ExpectRollback()
				})

				g.It("Should return an error", func() {
					err := repo.Store(ctx, chatMessage)

					Expect(err).ToNot(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetByID()", func() {
			var chatMessage *domain.ChatMessage

			g.BeforeEach(func() {
				chatMessage = &domain.ChatMessage{
					MessageID: 1,
					PlayerID:  "playerid",
					Platform:  "platform",
					ServerID:  1,
					Message:   "test chat message",
					Flagged:   true,
				}
			})

			g.Describe("Result found", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT * FROM ChatMessages")).WillReturnRows(sqlmock.NewRows(cols).
						AddRow(chatMessage.MessageID, chatMessage.PlayerID, chatMessage.Platform, chatMessage.ServerID,
							chatMessage.Message, chatMessage.Flagged, chatMessage.CreatedAt, chatMessage.ModifiedAt))
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetByID(ctx, chatMessage.MessageID)

					Expect(err).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct chat message", func() {
					msg, err := repo.GetByID(ctx, chatMessage.MessageID)

					Expect(err).To(BeNil())
					Expect(msg).To(Equal(chatMessage))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Result not found", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT * FROM ChatMessages")).WillReturnRows(sqlmock.NewRows(cols))
				})

				g.It("Should return domain.ErrNotFound", func() {
					_, err := repo.GetByID(ctx, chatMessage.MessageID)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})
		})
	})
}