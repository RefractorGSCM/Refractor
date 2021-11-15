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
	"Refractor/pkg/aeshelper"
	"Refractor/pkg/conf"
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"regexp"
	"strings"
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	var config = &conf.Config{
		EncryptionKey: strings.Repeat("a", 32),
	}

	var password = "password"
	var passwordEncrypted string
	encrypted, _ := aeshelper.Encrypt([]byte("password"), config.EncryptionKey)
	passwordEncrypted = string(encrypted)

	var cols = []string{"ServerID", "Game", "Name", "Address", "RCONPort", "RCONPassword", "Deactivated", "CreatedAt", "ModifiedAt"}

	g.Describe("Server Repo", func() {
		var repo domain.ServerRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo = NewServerRepo(db, zap.NewNop(), config)
		})

		g.Describe("Store()", func() {
			var mockServer *domain.Server

			g.BeforeEach(func() {
				mockServer = &domain.Server{
					ID:           1,
					Game:         "Mock",
					Name:         "Mock Server",
					Address:      "127.0.0.1",
					RCONPort:     "25575",
					RCONPassword: password,
					Deactivated:  false,
					CreatedAt:    time.Time{},
					ModifiedAt:   time.Time{},
				}

				mock.ExpectPrepare("INSERT INTO Servers")
			})

			g.Describe("Success", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery("INSERT INTO Servers").WillReturnRows(sqlmock.NewRows([]string{"ServerID"}).AddRow(int64(1)))

				})

				g.It("Should not return an error", func() {
					err := repo.Store(context.TODO(), mockServer)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should update the server to have the new ID", func() {
					_ = repo.Store(context.TODO(), mockServer)

					Expect(mockServer.ID).To(Equal(int64(1)))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Fail", func() {
				g.It("Should return an error on SQL error", func() {
					mock.ExpectQuery("INSERT INTO Servers").WillReturnError(fmt.Errorf(""))

					err := repo.Store(context.TODO(), mockServer)

					Expect(err).ToNot(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetByID()", func() {
			var mockServer *domain.Server

			g.BeforeEach(func() {
				mockServer = &domain.Server{
					ID:           1,
					Game:         "Mock",
					Name:         "Mock Server",
					Address:      "127.0.0.1",
					RCONPort:     "25575",
					RCONPassword: password,
					Deactivated:  false,
					CreatedAt:    time.Time{},
					ModifiedAt:   time.Time{},
				}
			})

			g.Describe("A result was found", func() {
				var mockRows *sqlmock.Rows

				g.BeforeEach(func() {
					mockRows = sqlmock.NewRows(cols).
						AddRow(mockServer.ID, mockServer.Game, mockServer.Name, mockServer.Address, mockServer.RCONPort,
							passwordEncrypted, mockServer.Deactivated, mockServer.CreatedAt, mockServer.ModifiedAt)
				})

				g.It("Should not return an error", func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Servers")).WillReturnRows(mockRows)

					_, err := repo.GetByID(context.TODO(), 1)

					Expect(err).To(BeNil())
				})

				g.It("Should return the correct rows scanned to a server object", func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Servers")).WillReturnRows(mockRows)

					server, err := repo.GetByID(context.TODO(), mockServer.ID)

					Expect(err).To(BeNil())
					Expect(server).ToNot(BeNil())
					Expect(server).To(Equal(mockServer))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No result found", func() {
				g.It("Should return domain.ErrNotFound if no results were found", func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Servers")).WillReturnRows(sqlmock.NewRows(cols))

					_, err := repo.GetByID(context.TODO(), 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("Deactivate()", func() {
			g.Describe("Target server exists", func() {
				g.BeforeEach(func() {
					mock.ExpectExec(regexp.QuoteMeta("UPDATE Servers SET Deactivated = TRUE")).
						WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.Deactivate(context.TODO(), 1)

					Expect(err).To(BeNil())
				})
			})

			g.Describe("Target server does not exist", func() {
				g.BeforeEach(func() {
					mock.ExpectExec(regexp.QuoteMeta("UPDATE Servers SET Deactivated = TRUE")).
						WillReturnResult(sqlmock.NewResult(0, 0))
				})

				g.It("Should return the error domain.ErrNotFound", func() {
					err := repo.Deactivate(context.TODO(), 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				})
			})
		})

		g.Describe("Update()", func() {
			g.BeforeEach(func() {
				mock.ExpectPrepare("UPDATE Servers SET")
			})

			g.Describe("Target row found", func() {
				var updatedServer *domain.Server
				var updateArgs domain.UpdateArgs

				g.BeforeEach(func() {
					us := &domain.Server{
						ID:           1,
						Game:         "Mock",
						Name:         "Updated Name",
						Address:      "127.0.0.1",
						RCONPort:     "25575",
						RCONPassword: password,
						Deactivated:  false,
						CreatedAt:    time.Time{},
						ModifiedAt:   time.Time{},
					}

					updateArgs = domain.UpdateArgs{
						"Name": "Updated Name",
					}

					updatedServer = us

					mockRows := sqlmock.NewRows(cols).
						AddRow(us.ID, us.Game, us.Name, us.Address, us.RCONPort,
							passwordEncrypted, us.Deactivated, us.CreatedAt, us.ModifiedAt)

					mock.ExpectQuery("UPDATE Servers SET").WillReturnRows(mockRows)
				})

				g.It("Should not return an error", func() {
					_, err := repo.Update(context.TODO(), updatedServer.ID, updateArgs)

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should scan and return the correct server", func() {
					updated, err := repo.Update(context.TODO(), updatedServer.ID, updateArgs)

					Expect(err).To(BeNil())
					Expect(updated).To(Equal(updatedServer))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Target row not found", func() {
				var updateArgs domain.UpdateArgs

				g.BeforeEach(func() {
					updateArgs = domain.UpdateArgs{
						"Name": "todo: name",
					}

					mock.ExpectQuery("UPDATE Servers SET").WillReturnError(sql.ErrNoRows)
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					_, err := repo.Update(context.TODO(), 5, updateArgs)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return a nil server", func() {
					g, _ := repo.Update(context.TODO(), 5, updateArgs)

					Expect(g).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		g.Describe("GetByGame()", func() {

			g.Describe("Results found", func() {
				var expected []*domain.Server

				g.BeforeEach(func() {
					expected = []*domain.Server{
						{
							ID:           1,
							Game:         "game1",
							Name:         "server 1",
							Address:      "127.0.0.1",
							RCONPort:     "1234",
							RCONPassword: password,
							Deactivated:  false,
							CreatedAt:    time.Time{},
							ModifiedAt:   time.Time{},
						},
						{
							ID:           2,
							Game:         "game1",
							Name:         "server 2",
							Address:      "127.0.0.1",
							RCONPort:     "1235",
							RCONPassword: password,
							Deactivated:  false,
							CreatedAt:    time.Time{},
							ModifiedAt:   time.Time{},
						},
						{
							ID:           3,
							Game:         "game1",
							Name:         "server 3",
							Address:      "127.0.0.1",
							RCONPort:     "1236",
							RCONPassword: password,
							Deactivated:  false,
							CreatedAt:    time.Time{},
							ModifiedAt:   time.Time{},
						},
					}

					rows := sqlmock.NewRows(cols)

					for _, s := range expected {
						rows.AddRow(s.ID, s.Game, s.Name, s.Address, s.RCONPort, passwordEncrypted, s.Deactivated,
							s.CreatedAt, s.ModifiedAt)
					}

					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Servers WHERE")).WillReturnRows(rows)
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetByGame(context.TODO(), "game1")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return the correct results", func() {
					got, err := repo.GetByGame(context.TODO(), "game1")

					Expect(err).To(BeNil())
					Expect(got).To(Equal(expected))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("No results found", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Servers WHERE")).WillReturnRows(sqlmock.NewRows(cols))
				})

				g.It("Should not return an error", func() {
					_, err := repo.GetByGame(context.TODO(), "game1")

					Expect(err).To(BeNil())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})

				g.It("Should return an empty slice", func() {
					got, err := repo.GetByGame(context.TODO(), "game1")

					Expect(err).To(BeNil())
					Expect(got).To(Equal([]*domain.Server{}))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			g.Describe("Database error", func() {
				g.BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Servers WHERE")).WillReturnError(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := repo.GetByGame(context.TODO(), "game1")

					Expect(err).ToNot(BeNil())
				})
			})
		})
	})
}
