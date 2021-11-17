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
	"github.com/lib/pq"
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
		"Repealed",
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
								i.Type, i.Reason, i.Duration, i.SystemAction, i.CreatedAt, i.ModifiedAt, i.Repealed))
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
							i.SystemAction, i.CreatedAt, i.ModifiedAt, i.Repealed)

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

		g.Describe("GetByPlayer()", func() {
			g.Describe("Infractions found", func() {
				var infractions []*domain.Infraction

				g.BeforeEach(func() {
					infractions = []*domain.Infraction{
						{
							InfractionID: 1,
							PlayerID:     "playerid",
							Platform:     "platform",
							UserID:       null.NewString("userid", true),
							ServerID:     1,
							Type:         domain.InfractionTypeMute,
							Reason:       null.NewString("Test mute reason", true),
							Duration:     null.NewInt(60, true),
							SystemAction: false,
							CreatedAt:    null.NewTime(time.Now(), true),
							ModifiedAt:   null.Time{},
						},
						{
							InfractionID: 2,
							PlayerID:     "playerid",
							Platform:     "platform",
							UserID:       null.NewString("userid2", true),
							ServerID:     2,
							Type:         domain.InfractionTypeKick,
							Reason:       null.NewString("Test kick reason", true),
							Duration:     null.NewInt(0, false),
							SystemAction: false,
							CreatedAt:    null.NewTime(time.Now(), true),
							ModifiedAt:   null.Time{},
						},
						{
							InfractionID: 3,
							PlayerID:     "playerid",
							Platform:     "platform",
							UserID:       null.NewString("userid2", true),
							ServerID:     2,
							Type:         domain.InfractionTypeWarning,
							Reason:       null.NewString("Test warn reason", true),
							Duration:     null.NewInt(0, false),
							SystemAction: false,
							CreatedAt:    null.NewTime(time.Now(), true),
							ModifiedAt:   null.Time{},
						},
						{
							InfractionID: 4,
							PlayerID:     "playerid",
							Platform:     "platform",
							UserID:       null.NewString("userid3", true),
							ServerID:     1,
							Type:         domain.InfractionTypeBan,
							Reason:       null.NewString("Test ban reason", true),
							Duration:     null.NewInt(1440, true),
							SystemAction: false,
							CreatedAt:    null.NewTime(time.Now(), true),
							ModifiedAt:   null.Time{},
						},
					}

					rows := sqlmock.NewRows(cols)
					for _, i := range infractions {
						rows.AddRow(i.InfractionID, i.PlayerID, i.Platform, i.UserID, i.ServerID, i.Type, i.Reason, i.Duration,
							i.SystemAction, i.CreatedAt, i.ModifiedAt, i.Repealed)
					}
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Infractions")).WillReturnRows(rows)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetByPlayer(ctx, "playerid", "platform")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct scanned results", func() {
					got, err := repo.GetByPlayer(ctx, "playerid", "platform")

					Expect(err).To(BeNil())
					Expect(got).To(Equal(infractions))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No infractions found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Infractions")).WillReturnRows(sqlmock.NewRows(cols))
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					_, err := repo.GetByPlayer(ctx, "playerid", "platform")

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Infractions")).WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := repo.GetByPlayer(ctx, "playerid", "platform")

					Expect(err).ToNot(BeNil())
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
							ui.Duration, ui.SystemAction, ui.CreatedAt, ui.ModifiedAt, ui.Repealed))
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

		g.Describe("Search()", func() {
			var cols = []string{"InfractionID", "PlayerID", "Platform", "UserID", "ServerID", "Type", "Reason", "Duration",
				"SystemAction", "CreatedAt", "ModifiedAt", "Repealed", "StaffName"}

			g.Describe("Results found", func() {
				var results []*domain.Infraction

				g.BeforeEach(func() {
					results = []*domain.Infraction{
						{
							InfractionID: 1,
							PlayerID:     "playerid",
							Platform:     "platform",
							UserID:       null.NewString("userid", true),
							ServerID:     1,
							Type:         domain.InfractionTypeWarning,
							Reason:       null.NewString("reason", true),
							Duration:     null.Int{},
							SystemAction: true,
							CreatedAt:    null.Time{},
							ModifiedAt:   null.Time{},
							IssuerName:   "username",
						},
						{
							InfractionID: 2,
							PlayerID:     "playerid2",
							Platform:     "platform2",
							UserID:       null.NewString("userid", true),
							ServerID:     1,
							Type:         domain.InfractionTypeBan,
							Reason:       null.NewString("reason", true),
							Duration:     null.NewInt(60, true),
							SystemAction: false,
							CreatedAt:    null.Time{},
							ModifiedAt:   null.Time{},
							IssuerName:   "username",
						},
						{
							InfractionID: 1,
							PlayerID:     "playerid3",
							Platform:     "platform3",
							UserID:       null.NewString("userid", true),
							ServerID:     1,
							Type:         domain.InfractionTypeKick,
							Reason:       null.NewString("reason", true),
							Duration:     null.Int{},
							SystemAction: false,
							CreatedAt:    null.Time{},
							ModifiedAt:   null.Time{},
							IssuerName:   "username",
						},
					}

					rows := sqlmock.NewRows(cols)

					for _, i := range results {
						rows.AddRow(i.InfractionID, i.PlayerID, i.Platform, i.UserID, i.ServerID, i.Type, i.Reason,
							i.Duration, i.SystemAction, i.CreatedAt, i.ModifiedAt, i.Repealed, i.IssuerName)
					}

					mock.ExpectQuery(regexp.QuoteMeta("SELECT res.*, um.Username AS StaffName FROM (")).WillReturnRows(rows)
					mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(1) AS Count FROM Infractions")).WillReturnRows(sqlmock.NewRows([]string{"count"}).
						AddRow(1000))
				})

				g.It("Should not return an error", func() {
					_, _, err := repo.Search(ctx, domain.FindArgs{}, nil, 10, 0)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the expected results", func() {
					_, got, err := repo.Search(ctx, domain.FindArgs{}, nil, 10, 0)

					Expect(err).To(BeNil())
					Expect(got).To(Equal(results))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct total results", func() {
					total, _, err := repo.Search(ctx, domain.FindArgs{}, nil, 10, 0)

					Expect(err).To(BeNil())
					Expect(total).To(Equal(1000))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No results found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT res.*, um.Username AS StaffName FROM (")).WillReturnRows(
						sqlmock.NewRows(cols))
				})

				g.It("Should not return an error", func() {
					_, _, err := repo.Search(ctx, domain.FindArgs{}, nil, 10, 0)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return an empty array and a count of 0", func() {
					total, got, err := repo.Search(ctx, domain.FindArgs{}, nil, 10, 0)

					Expect(err).To(BeNil())
					Expect(got).To(Equal([]*domain.Infraction{}))
					Expect(total).To(Equal(0))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT res.*, um.Username AS StaffName FROM (")).WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, _, err := repo.Search(ctx, domain.FindArgs{}, nil, 10, 0)

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetLinkedChatMessages()", func() {
			var expected []*domain.ChatMessage

			g.Describe("Linked messages found", func() {
				g.BeforeEach(func() {
					expected = []*domain.ChatMessage{
						{
							MessageID: 1,
							PlayerID:  "playerid1",
							Platform:  "platform",
							ServerID:  1,
							Message:   "msg1",
							Flagged:   true,
						},
						{
							MessageID: 2,
							PlayerID:  "playerid2",
							Platform:  "platform",
							ServerID:  1,
							Message:   "msg2",
							Flagged:   true,
						},
						{
							MessageID: 3,
							PlayerID:  "playerid3",
							Platform:  "platform",
							ServerID:  1,
							Message:   "msg33",
							Flagged:   true,
						},
					}

					rows := sqlmock.NewRows([]string{
						"MessageID", "PlayerID", "Platform", "ServerID", "Message", "Flagged", "CreatedAt", "ModifiedAt",
					})

					for _, r := range expected {
						rows.AddRow(r.MessageID, r.PlayerID, r.Platform, r.ServerID, r.Message, r.Flagged, r.CreatedAt, r.ModifiedAt)
					}

					mock.ExpectQuery(regexp.QuoteMeta("SELECT cm.MessageID")).WillReturnRows(rows)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetLinkedChatMessages(ctx, 1)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct results", func() {
					got, err := repo.GetLinkedChatMessages(ctx, 1)

					Expect(err).To(BeNil())
					Expect(got).To(Equal(expected))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No linked messages found", func() {
				g.BeforeEach(func() {
					expected = []*domain.ChatMessage{}

					rows := sqlmock.NewRows([]string{
						"MessageID", "PlayerID", "Platform", "ServerID", "Message", "Flagged", "CreatedAt", "ModifiedAt",
					})

					mock.ExpectQuery(regexp.QuoteMeta("SELECT cm.MessageID")).WillReturnRows(rows)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetLinkedChatMessages(ctx, 1)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return an empty slice", func() {
					got, err := repo.GetLinkedChatMessages(ctx, 1)

					Expect(err).To(BeNil())
					Expect(got).To(Equal(expected))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					expected = []*domain.ChatMessage{}

					mock.ExpectQuery(regexp.QuoteMeta("SELECT cm.MessageID")).WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetLinkedChatMessages(ctx, 1)

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("LinkChatMessages()", func() {
			g.Describe("Successful link", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO InfractionChatMessages").WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.LinkChatMessages(ctx, 1, 1)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Foreign key error (chat message or infraction ids are invalid)", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO InfractionChatMessages").WillReturnError(pq.Error{Code: "23503"}) // code for postgres FK errors
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					err := repo.LinkChatMessages(ctx, 1, 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Unexpected database error", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO InfractionChatMessages").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.LinkChatMessages(ctx, 1, 1)

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("LinkChatMessages()", func() {
			g.Describe("Successful unlink", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM InfractionChatMessages").WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.UnlinkChatMessages(ctx, 1, 1)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Link not found", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM InfractionChatMessages").WillReturnResult(sqlmock.NewResult(0, 0))
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					err := repo.UnlinkChatMessages(ctx, 1, 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Unexpected database error", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM InfractionChatMessages").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.UnlinkChatMessages(ctx, 1, 1)

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetMostSignificantInfraction()", func() {
			g.Describe("Results found", func() {
				var results []*domain.Infraction

				g.BeforeEach(func() {
					results = []*domain.Infraction{
						{
							InfractionID: 1,
							PlayerID:     "playerid1",
							Platform:     "testplatform",
							UserID:       null.String{},
							ServerID:     1,
							Type:         domain.InfractionTypeBan,
							Reason:       null.NewString("test reason", true),
							Duration:     null.NewInt(1440, true),
							SystemAction: false,
							CreatedAt:    null.Time{},
							ModifiedAt:   null.Time{},
							Repealed:     false,
						},
						{
							InfractionID: 2,
							PlayerID:     "playerid1",
							Platform:     "testplatform",
							UserID:       null.String{},
							ServerID:     2,
							Type:         domain.InfractionTypeBan,
							Reason:       null.NewString("test reason", true),
							Duration:     null.NewInt(1000, true),
							SystemAction: false,
							CreatedAt:    null.Time{},
							ModifiedAt:   null.Time{},
							Repealed:     false,
						},
					}

					rows := sqlmock.NewRows(cols)

					for _, i := range results {
						rows.AddRow(i.InfractionID, i.PlayerID, i.Platform, i.UserID, i.ServerID, i.Type, i.Reason,
							i.Duration, i.SystemAction, i.CreatedAt, i.ModifiedAt, i.Repealed)
					}

					mock.ExpectQuery(regexp.QuoteMeta("select * from infractions")).WillReturnRows(rows)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetMostSignificantInfraction(ctx, domain.InfractionTypeBan, "testplatform", "playerid1")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct infraction", func() {
					expected := results[0]
					got, err := repo.GetMostSignificantInfraction(ctx, domain.InfractionTypeBan, "testplatform", "playerid1")

					Expect(err).To(BeNil())
					Expect(got).To(Equal(expected))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No results found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("select * from infractions")).WillReturnRows(sqlmock.NewRows(cols))
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetMostSignificantInfraction(ctx, domain.InfractionTypeBan, "testplatform", "playerid1")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return nil", func() {
					got, err := repo.GetMostSignificantInfraction(ctx, domain.InfractionTypeBan, "testplatform", "playerid1")

					Expect(err).To(BeNil())
					Expect(got).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("select * from infractions")).WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := repo.GetMostSignificantInfraction(ctx, domain.InfractionTypeBan, "testplatform", "playerid1")

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return nil", func() {
					got, err := repo.GetMostSignificantInfraction(ctx, domain.InfractionTypeBan, "testplatform", "playerid1")

					Expect(err).ToNot(BeNil())
					Expect(got).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetPlayerTotalInfractions()", func() {
			g.Describe("Infraction count returned successfully", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(1) FROM Infractions")).WillReturnRows(sqlmock.
						NewRows([]string{"Count"}).AddRow(13287))
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetPlayerTotalInfractions(ctx, "platform", "playerid")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct count", func() {
					count, err := repo.GetPlayerTotalInfractions(ctx, "platform", "playerid")

					Expect(err).To(BeNil())
					Expect(count).To(Equal(13287))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No infractions found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(1) FROM Infractions")).WillReturnRows(sqlmock.
						NewRows([]string{"Count"}).AddRow(0))
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetPlayerTotalInfractions(ctx, "platform", "playerid")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return zero as the count", func() {
					count, err := repo.GetPlayerTotalInfractions(ctx, "platform", "playerid")

					Expect(err).To(BeNil())
					Expect(count).To(Equal(0))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(1) FROM")).WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := repo.GetPlayerTotalInfractions(ctx, "platform", "playerid")

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})
	})
}
