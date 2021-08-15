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

	var cols = []string{"GroupID", "Name", "Color", "Position", "Permissions", "CreatedAt", "ModifiedAt"}

	g.Describe("Store()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())

			mock.ExpectPrepare("INSERT INTO Groups")
		})

		g.After(func() {
			_ = db.Close()
		})

		g.Describe("Success", func() {
			g.It("Should not return an error", func() {
				mock.ExpectQuery("INSERT INTO Groups").WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

				group := &domain.Group{Name: "Test"}
				err := repo.Store(context.TODO(), group)

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should update the group to have the new ID", func() {
				mock.ExpectQuery("INSERT INTO Groups").WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

				group := &domain.Group{Name: "Test"}
				_ = repo.Store(context.TODO(), group)

				Expect(group.ID).To(Equal(int64(1)))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("Fail", func() {
			g.It("Should return an error on SQL error", func() {
				mock.ExpectQuery("INSERT INTO Groups").WillReturnError(fmt.Errorf("err"))

				group := &domain.Group{Name: "Test"}
				err := repo.Store(context.TODO(), group)

				Expect(err).ToNot(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})
	})

	g.Describe("GetByID()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmocker. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())
		})

		g.After(func() {
			_ = db.Close()
		})

		g.Describe("A result was found", func() {
			var mockGroup *domain.Group
			var mockRows *sqlmock.Rows

			g.BeforeEach(func() {
				mockGroup = &domain.Group{
					ID:          1,
					Name:        "Mock",
					Color:       763763,
					Position:    4,
					Permissions: "347632748",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				}

				mockRows = sqlmock.NewRows(cols).
					AddRow(mockGroup.ID, mockGroup.Name, mockGroup.Color, mockGroup.Position, mockGroup.Permissions,
						mockGroup.CreatedAt, mockGroup.ModifiedAt)
			})

			g.It("Should not return an error", func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Groups")).WillReturnRows(mockRows)

				_, err := repo.GetByID(context.TODO(), mockGroup.ID)

				Expect(err).To(BeNil())
			})

			g.It("Should return the correct rows scanned to a group object", func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Groups")).WillReturnRows(mockRows)

				group, _ := repo.GetByID(context.TODO(), mockGroup.ID)

				Expect(group).ToNot(BeNil())
				Expect(group).To(Equal(mockGroup))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("No result found", func() {
			g.It("Should return domain.ErrNotFound", func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Groups")).WillReturnRows(sqlmock.NewRows(cols))
				_, err := repo.GetByID(context.TODO(), 1)

				Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})
	})

	g.Describe("GetAll()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())
		})

		g.After(func() {
			_ = db.Close()
		})

		g.Describe("Results were found", func() {
			var mockGroups []*domain.Group
			var mockRows *sqlmock.Rows

			g.BeforeEach(func() {
				mockGroups = append(mockGroups, &domain.Group{
					ID:          1,
					Name:        "Mock",
					Color:       763763,
					Position:    4,
					Permissions: "347632748",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				})

				mockGroups = append(mockGroups, &domain.Group{
					ID:          2,
					Name:        "Mock 2",
					Color:       76123763,
					Position:    4,
					Permissions: "347876434632748",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				})

				mockGroups = append(mockGroups, &domain.Group{
					ID:          3,
					Name:        "Mock 3",
					Color:       76365763,
					Position:    4,
					Permissions: "3465367632748",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				})

				mockRows = sqlmock.NewRows(cols)

				for _, mockGroup := range mockGroups {
					mockRows.AddRow(mockGroup.ID, mockGroup.Name, mockGroup.Color, mockGroup.Position, mockGroup.Permissions,
						mockGroup.CreatedAt, mockGroup.ModifiedAt)
				}
			})

			g.It("Should not return an error", func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Groups")).WillReturnRows(mockRows)

				_, err := repo.GetAll(context.TODO())

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return the results scanned in an array", func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Groups")).WillReturnRows(mockRows)

				results, _ := repo.GetAll(context.TODO())

				Expect(results).To(Equal(results))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("No results found", func() {
			g.It("Should not return an error", func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Groups")).WillReturnRows(sqlmock.NewRows(cols))

				_, err := repo.GetAll(context.TODO())

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return just the base group", func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Groups")).WillReturnRows(sqlmock.NewRows(cols))

				res, _ := repo.GetAll(context.TODO())

				Expect(res).To(Equal([]*domain.Group{}))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})
	})

	g.Describe("GetUserGroups()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB
		var cols []string

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())

			cols = []string{"GroupID", "Name", "Color", "Position", "Permissions", "CreatedAt", "ModifiedAt"}
		})

		g.Describe("NewUser groups were found", func() {
			var groups []*domain.Group

			g.Before(func() {
				groups = append(groups, &domain.Group{
					ID:          1,
					Name:        "Group 1",
					Color:       1234,
					Position:    1,
					Permissions: "2657643576",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				})

				groups = append(groups, &domain.Group{
					ID:          2,
					Name:        "Group 2",
					Color:       54533,
					Position:    2,
					Permissions: "763456333",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				})
			})

			g.BeforeEach(func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT g.* FROM UserGroups")).WillReturnRows(sqlmock.NewRows(cols).
					AddRow(groups[0].ID, groups[0].Name, groups[0].Color, groups[0].Position, groups[0].Permissions, groups[0].CreatedAt, groups[0].ModifiedAt).
					AddRow(groups[1].ID, groups[1].Name, groups[1].Color, groups[1].Position, groups[1].Permissions, groups[1].CreatedAt, groups[1].ModifiedAt))
			})

			g.It("Should not return an error", func() {
				_, err := repo.GetUserGroups(context.TODO(), "userid")

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should scan and return user groups", func() {
				foundGroups, _ := repo.GetUserGroups(context.TODO(), "userid")

				Expect(foundGroups).To(Equal(groups))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("No user groups were found", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT g.* FROM UserGroups")).WillReturnRows(sqlmock.NewRows(cols))
			})

			g.It("Should return domain.ErrNotFound error", func() {
				_, err := repo.GetUserGroups(context.TODO(), "userid")

				Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return nil", func() {
				foundGroups, _ := repo.GetUserGroups(context.TODO(), "userid")

				Expect(foundGroups).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})
	})

	g.Describe("GetUserOverrides()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB
		var cols []string

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())

			cols = []string{"AllowOverrides", "DenyOverrides"}
		})

		g.Describe("NewUser overrides were found", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery("SELECT AllowOverrides, DenyOverrides FROM UserOverrides").WillReturnRows(
					sqlmock.NewRows(cols).AddRow("1234", "12345"))
			})

			g.It("Should not return an error", func() {
				_, err := repo.GetUserOverrides(context.TODO(), "testuserid")

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return an Overrides struct with the scanned values", func() {
				expected := &domain.Overrides{
					AllowOverrides: "1234",
					DenyOverrides:  "12345",
				}

				overrides, _ := repo.GetUserOverrides(context.TODO(), "testuserid")

				Expect(overrides).To(Equal(expected))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("No overrides were found", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery("SELECT AllowOverrides, DenyOverrides FROM UserOverrides").WillReturnRows(
					sqlmock.NewRows(cols))
			})

			g.It("Should return domain.ErrNotFound", func() {
				_, err := repo.GetUserOverrides(context.TODO(), "testuserid")

				Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return nil", func() {
				overrides, _ := repo.GetUserOverrides(context.TODO(), "testuserid")

				Expect(overrides).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})
	})

	g.Describe("GetServerOverrides()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB
		var cols []string

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())

			cols = []string{"AllowOverrides", "DenyOverrides"}
		})

		g.Describe("Server overrides were found", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery("SELECT AllowOverrides, DenyOverrides FROM ServerGroups").WillReturnRows(
					sqlmock.NewRows(cols).AddRow("1234", "12345"))
			})

			g.It("Should not return an error", func() {
				_, err := repo.GetServerOverrides(context.TODO(), 1, 1)

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return an Overrides struct with the scanned values", func() {
				expected := &domain.Overrides{
					AllowOverrides: "1234",
					DenyOverrides:  "12345",
				}

				overrides, _ := repo.GetServerOverrides(context.TODO(), 1, 1)

				Expect(overrides).To(Equal(expected))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("No overrides were found", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery("SELECT AllowOverrides, DenyOverrides FROM ServerGroups").WillReturnRows(
					sqlmock.NewRows(cols))
			})

			g.It("Should return domain.ErrNotFound", func() {
				_, err := repo.GetServerOverrides(context.TODO(), 1, 1)

				Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return nil", func() {
				overrides, _ := repo.GetServerOverrides(context.TODO(), 1, 1)

				Expect(overrides).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})
	})

	g.Describe("GetServerOverridesAllGroups()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB
		var cols []string

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())

			cols = []string{"GroupID", "AllowOverrides", "DenyOverrides"}
		})

		g.Describe("Server overrides were found", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery("SELECT GroupID, AllowOverrides, DenyOverrides FROM ServerGroups").WillReturnRows(
					sqlmock.NewRows(cols).AddRow(1, "1234", "12345").AddRow(2, "543", "42523"))
			})

			g.It("Should not return an error", func() {
				_, err := repo.GetServerOverridesAllGroups(context.TODO(), 1)

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return an array of Overrides structs with the scanned values", func() {
				expected := []*domain.Overrides{
					{
						GroupID:        1,
						AllowOverrides: "1234",
						DenyOverrides:  "12345",
					},
					{
						GroupID:        2,
						AllowOverrides: "543",
						DenyOverrides:  "42523",
					},
				}

				overrides, _ := repo.GetServerOverridesAllGroups(context.TODO(), 1)

				Expect(overrides).To(Equal(expected))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("No overrides were found", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery("SELECT GroupID, AllowOverrides, DenyOverrides FROM ServerGroups").WillReturnRows(
					sqlmock.NewRows(cols))
			})

			g.It("Should return domain.ErrNotFound", func() {
				_, err := repo.GetServerOverridesAllGroups(context.TODO(), 1)

				Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return nil", func() {
				overrides, _ := repo.GetServerOverridesAllGroups(context.TODO(), 1)

				Expect(overrides).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})
	})

	g.Describe("Delete()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())
		})

		g.Describe("Target group exists", func() {
			g.BeforeEach(func() {
				mock.ExpectExec("DELETE FROM Groups").WillReturnResult(sqlmock.NewResult(0, 1))
			})

			g.It("Should not return an error", func() {
				err := repo.Delete(context.TODO(), 1)

				Expect(err).To(BeNil())
			})
		})

		g.Describe("Target group does not exist", func() {
			g.BeforeEach(func() {
				mock.ExpectExec("DELETE FROM Groups").WillReturnResult(sqlmock.NewResult(0, 0))
			})

			g.It("Should return the error domain.ErrNotFound", func() {
				err := repo.Delete(context.TODO(), 1)

				Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
			})
		})
	})

	g.Describe("Update()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())

			mock.ExpectPrepare("UPDATE Groups SET")
		})

		g.Describe("Target row found", func() {
			var updatedGroup *domain.Group
			var updateArgs domain.UpdateArgs

			g.BeforeEach(func() {
				ug := &domain.Group{
					ID:          2,
					Name:        "Updated Group",
					Color:       0xcecece,
					Position:    5,
					Permissions: "482736",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				}

				updateArgs = domain.UpdateArgs{
					"Permissions": ug.Permissions,
				}

				updatedGroup = ug

				mock.ExpectQuery("UPDATE Groups SET").WillReturnRows(sqlmock.NewRows(cols).
					AddRow(ug.ID, ug.Name, ug.Color, ug.Position, ug.Permissions, ug.CreatedAt, ug.ModifiedAt))
			})

			g.It("Should not return an error", func() {
				_, err := repo.Update(context.TODO(), updatedGroup.ID, updateArgs)

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should scan and return the correct group", func() {
				updated, err := repo.Update(context.TODO(), updatedGroup.ID, updateArgs)

				Expect(err).To(BeNil())
				Expect(updated).To(Equal(updatedGroup))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("Target row not found", func() {
			var updateArgs domain.UpdateArgs

			g.BeforeEach(func() {
				updateArgs = domain.UpdateArgs{
					"Permissions": "6732636748",
				}

				mock.ExpectQuery("UPDATE Groups SET").WillReturnError(sql.ErrNoRows)
			})

			g.It("Should return a domain.ErrNotFound error", func() {
				_, err := repo.Update(context.TODO(), 5, updateArgs)

				Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should return a nil group", func() {
				g, _ := repo.Update(context.TODO(), 5, updateArgs)

				Expect(g).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})
	})

	g.Describe("Reorder()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())
		})

		g.Describe("Reorder successful", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery("SELECT reorder_groups").WillReturnRows(sqlmock.NewRows([]string{}).AddRow())
			})

			g.It("Should not return an error", func() {
				err := repo.Reorder(context.TODO(), []*domain.GroupReorderInfo{})

				Expect(err).To(BeNil())
			})
		})

		g.Describe("Reorder error", func() {
			g.BeforeEach(func() {
				mock.ExpectQuery("SELECT reorder_groups").WillReturnError(fmt.Errorf(""))
			})

			g.It("Should return an error", func() {
				err := repo.Reorder(context.TODO(), []*domain.GroupReorderInfo{})

				Expect(err).ToNot(BeNil())
			})
		})
	})

	g.Describe("AddUserGroup()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())
		})

		g.Describe("Target group exists", func() {
			g.Describe("User does not yet have this group", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO UserGroups").WillReturnResult(sqlmock.NewResult(0, 1))
				})

				g.It("Should not return an error", func() {
					err := repo.AddUserGroup(context.TODO(), "userid", 1)

					Expect(err).To(BeNil())
				})
			})

			g.Describe("User already has this group", func() {
				g.BeforeEach(func() {
					mock.ExpectExec("INSERT INTO UserGroups").WillReturnError(&pq.Error{Constraint: "usergroups_pkey"})
				})

				g.It("Should return domain.ErrConflict error", func() {
					err := repo.AddUserGroup(context.TODO(), "userid", 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrConflict))
				})
			})
		})

		g.Describe("Target group does not exist", func() {
			g.BeforeEach(func() {
				mock.ExpectExec("INSERT INTO UserGroups").WillReturnError(fmt.Errorf("err"))
			})

			g.It("Should return an error", func() {
				err := repo.AddUserGroup(context.TODO(), "userid", 1)

				Expect(err).ToNot(BeNil())
			})
		})
	})

	g.Describe("RemoveUserGroup()", func() {
		var repo domain.GroupRepo
		var mock sqlmock.Sqlmock
		var db *sql.DB

		g.BeforeEach(func() {
			var err error

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Could not create new sqlmock instance. Error: %v", err)
			}

			repo, _ = NewGroupRepo(db, zap.NewNop())
		})

		g.Describe("Target group exists", func() {
			g.BeforeEach(func() {
				mock.ExpectExec("DELETE FROM UserGroups").WillReturnResult(sqlmock.NewResult(0, 1))
			})

			g.It("Should not return an error", func() {
				err := repo.RemoveUserGroup(context.TODO(), "userid", 1)

				Expect(err).To(BeNil())
			})
		})

		g.Describe("Target group does not exist", func() {
			g.BeforeEach(func() {
				mock.ExpectExec("DELETE FROM UserGroups").WillReturnResult(sqlmock.NewResult(0, 0))
			})

			g.It("Should not return an error", func() {
				err := repo.RemoveUserGroup(context.TODO(), "userid", 1)

				Expect(err).To(BeNil())
			})
		})
	})
}
