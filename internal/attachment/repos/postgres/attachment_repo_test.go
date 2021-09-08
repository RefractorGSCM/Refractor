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

	var cols = []string{"AttachmentID", "InfractionID", "URL", "Note"}

	g.Describe("Postgres Attachment Repo", func() {
		var repo domain.AttachmentRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB
		var ctx context.Context

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo = NewAttachmentRepo(db, zap.NewNop())
			ctx = context.TODO()
		})

		g.AfterEach(func() {
			_ = db.Close()
		})

		g.Describe("Store()", func() {
			g.Describe("Successful store", func() {
				var expected *domain.Attachment

				g.BeforeEach(func() {
					expected = &domain.Attachment{
						InfractionID: 1,
						URL:          "test.com/img.png",
						Note:         "Test note",
					}

					mock.ExpectPrepare("INSERT INTO Attachments")
					mock.ExpectQuery("INSERT INTO Attachments").WillReturnRows(
						sqlmock.NewRows([]string{"InfractionID"}).
							AddRow(int64(10)))
				})

				g.It("Should not return an error", func() {
					err := repo.Store(ctx, expected)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should update the passed in attachment to have the new ID", func() {
					err := repo.Store(ctx, expected)

					Expect(err).To(BeNil())
					Expect(expected.AttachmentID).To(Equal(int64(10)))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectPrepare("INSERT INTO Attachments")
					mock.ExpectQuery("INSERT INTO Attachments").WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := repo.Store(ctx, &domain.Attachment{})

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetByInfraction()", func() {
			g.Describe("Attachments found", func() {
				var mockAttachments []*domain.Attachment
				var mockRows *sqlmock.Rows

				g.BeforeEach(func() {
					mockAttachments = []*domain.Attachment{
						{
							AttachmentID: 1,
							InfractionID: 1,
							URL:          "https://test.com/img.png",
							Note:         "Attachment 1",
						},
						{
							AttachmentID: 2,
							InfractionID: 1,
							URL:          "https://test2.com/img.png",
							Note:         "Attachment 2",
						},
						{
							AttachmentID: 3,
							InfractionID: 1,
							URL:          "https://test3.com/img.png",
							Note:         "Attachment 3",
						},
					}

					mockRows = sqlmock.NewRows(cols)

					for _, attachment := range mockAttachments {
						mockRows.AddRow(attachment.AttachmentID, attachment.InfractionID, attachment.URL, attachment.Note)
					}

					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Attachments")).WillReturnRows(mockRows)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetByInfraction(ctx, 1)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct rows scanned into attachments", func() {
					found, err := repo.GetByInfraction(ctx, 1)

					Expect(err).To(BeNil())
					Expect(found).To(Equal(mockAttachments))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No attachments found found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Attachments")).WillReturnRows(sqlmock.NewRows(cols))
				})

				g.It("Should return domain.ErrNotFound error", func() {
					_, err := repo.GetByInfraction(ctx, 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Delete()", func() {
			g.Describe("Target attachment exists", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM Attachments").WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.Delete(ctx, 1)

					Expect(err).To(BeNil())
				})
			})

			g.Describe("Target attachment does not exist", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("DELETE FROM Attachments").WillReturnResult(sqlmock.NewResult(0, 0))
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					err := repo.Delete(ctx, 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				})
			})
		})
	})
}
