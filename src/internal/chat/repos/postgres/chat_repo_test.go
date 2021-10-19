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
					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT MessageID, PlayerID, Platform")).WillReturnRows(sqlmock.NewRows(cols).
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
					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT MessageID, PlayerID, Platform")).WillReturnRows(sqlmock.NewRows(cols))
				})

				g.It("Should return domain.ErrNotFound", func() {
					_, err := repo.GetByID(ctx, chatMessage.MessageID)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetRecentByServer()", func() {
			var rows *sqlmock.Rows

			g.Describe("Results found", func() {
				var messages []*domain.ChatMessage

				g.BeforeEach(func() {
					messages = []*domain.ChatMessage{
						{
							MessageID: 1,
							PlayerID:  "playerid",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 1",
							Flagged:   false,
						}, {
							MessageID: 2,
							PlayerID:  "playerid2",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 2",
							Flagged:   false,
						}, {
							MessageID: 3,
							PlayerID:  "playerid3",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 3",
							Flagged:   false,
						}, {
							MessageID: 4,
							PlayerID:  "playerid4",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 4",
							Flagged:   false,
						}, {
							MessageID: 5,
							PlayerID:  "playerid5",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 5",
							Flagged:   false,
						}, {
							MessageID: 6,
							PlayerID:  "playerid6",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 6",
							Flagged:   false,
						}, {
							MessageID: 7,
							PlayerID:  "playerid7",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 7",
							Flagged:   false,
						}, {
							MessageID: 8,
							PlayerID:  "playerid8",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 8",
							Flagged:   false,
						}, {
							MessageID: 9,
							PlayerID:  "playerid9",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 9",
							Flagged:   false,
						}, {
							MessageID: 10,
							PlayerID:  "playerid10",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 10",
							Flagged:   false,
						},
					}

					rows = sqlmock.NewRows(cols)

					for _, msg := range messages {
						rows.AddRow(msg.MessageID, msg.PlayerID, msg.Platform, msg.ServerID, msg.Message, msg.Flagged, msg.CreatedAt, msg.ModifiedAt)
					}

					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT MessageID, PlayerID, Platform")).WillReturnRows(rows)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetRecentByServer(ctx, 1, 10)

					Expect(err).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct results", func() {
					results, err := repo.GetRecentByServer(ctx, 1, 10)

					Expect(err).To(BeNil())
					Expect(results).To(Equal(messages))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No results found", func() {
				g.BeforeEach(func() {
					rows = sqlmock.NewRows(cols)

					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT MessageID, PlayerID, Platform")).WillReturnRows(rows)
				})

				g.It("Should return domain.ErrNotFound error", func() {
					_, err := repo.GetRecentByServer(ctx, 1, 10)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Search()", func() {
			g.Describe("Results found", func() {
				var results []*domain.ChatMessage

				g.BeforeEach(func() {
					results = []*domain.ChatMessage{
						{
							MessageID: 1,
							PlayerID:  "playerid1",
							Platform:  "platform",
							ServerID:  1,
							Message:   "test message 1",
							Flagged:   false,
						},
						{
							MessageID: 2,
							PlayerID:  "playerid2",
							Platform:  "platform",
							ServerID:  2,
							Message:   "test message 2",
							Flagged:   false,
						},
						{
							MessageID: 3,
							PlayerID:  "playerid3",
							Platform:  "platform",
							ServerID:  3,
							Message:   "test message 3",
							Flagged:   false,
						},
						{
							MessageID: 4,
							PlayerID:  "playerid4",
							Platform:  "platform",
							ServerID:  4,
							Message:   "test message 4",
							Flagged:   false,
						},
						{
							MessageID: 5,
							PlayerID:  "playerid5",
							Platform:  "platform",
							ServerID:  5,
							Message:   "test message 1",
							Flagged:   false,
						},
					}

					rows := sqlmock.NewRows(cols)

					for _, msg := range results {
						rows.AddRow(msg.MessageID, msg.PlayerID, msg.Platform, msg.ServerID, msg.Message, msg.Flagged,
							msg.CreatedAt, msg.ModifiedAt)
					}

					mockRepo.ExpectQuery("SELECT cm.MessageID, cm.PlayerID, cm.Platform").WillReturnRows(rows)
					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(1) AS Count FROM ChatMessages")).WillReturnRows(sqlmock.NewRows([]string{"Count"}).
						AddRow(10))
				})

				g.It("Should not return an error", func() {
					_, _, err := repo.Search(ctx, domain.FindArgs{}, []int64{}, 5, 0)

					Expect(err).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the expected results", func() {
					_, got, err := repo.Search(ctx, domain.FindArgs{}, []int64{}, 5, 0)

					Expect(err).To(BeNil())
					Expect(got).To(Equal(results))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct number of total results", func() {
					total, _, err := repo.Search(ctx, domain.FindArgs{}, []int64{}, 10, 0)

					Expect(err).To(BeNil())
					Expect(total).To(Equal(10))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No results found", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery("SELECT cm.MessageID, cm.PlayerID, cm.Platform").WillReturnRows(sqlmock.NewRows(cols))
				})

				g.It("Should not return an error", func() {
					_, _, err := repo.Search(ctx, domain.FindArgs{}, []int64{}, 10, 0)

					Expect(err).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return an empty array and a count of 0", func() {
					total, got, err := repo.Search(ctx, domain.FindArgs{}, []int64{}, 10, 0)

					Expect(err).To(BeNil())
					Expect(got).To(Equal([]*domain.ChatMessage{}))
					Expect(total).To(Equal(0))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery("SELECT cm.MessageID, cm.PlayerID, cm.Platform").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, _, err := repo.Search(ctx, domain.FindArgs{}, []int64{}, 10, 0)

					Expect(err).ToNot(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("TSQuery error", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery("SELECT cm.MessageID, cm.PlayerID, cm.Platform").WillReturnError(fmt.Errorf("syntax error in tsquery"))
				})

				g.It("Should return a domain.ErrInvalidQuery error", func() {
					_, _, err := repo.Search(ctx, domain.FindArgs{}, []int64{}, 10, 0)

					Expect(errors.Cause(err)).To(Equal(domain.ErrInvalidQuery))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetFlaggedMessageCount()", func() {
			g.Describe("Count fetched", func() {
				g.BeforeEach(func() {
					mockRepo.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(1) FROM ChatMessages WHERE")).
						WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1823))
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetFlaggedMessageCount(ctx)

					Expect(err).To(BeNil())
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct amount", func() {
					count, err := repo.GetFlaggedMessageCount(ctx)

					Expect(err).To(BeNil())
					Expect(count).To(Equal(1823))
					Expect(mockRepo.ExpectationsWereMet()).To(BeNil())
				})
			})
		})
	})
}
