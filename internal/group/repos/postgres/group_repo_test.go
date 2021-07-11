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

			repo = NewGroupRepo(db, zap.NewNop())

			mock.ExpectPrepare("INSERT INTO Groups")
		})

		g.After(func() {
			_ = db.Close()
		})

		g.Describe("Success", func() {
			g.It("Should not return an error", func() {
				mock.ExpectExec("INSERT INTO Groups").WillReturnResult(sqlmock.NewResult(1, 1))

				group := &domain.Group{Name: "Test"}
				err := repo.Store(context.TODO(), group)

				Expect(err).To(BeNil())
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})

			g.It("Should update the group to have the new ID", func() {
				mock.ExpectExec("INSERT INTO Groups").WillReturnResult(sqlmock.NewResult(1, 1))

				group := &domain.Group{Name: "Test"}
				_ = repo.Store(context.TODO(), group)

				Expect(group.ID).To(Equal(int64(1)))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		g.Describe("Fail", func() {
			g.It("Should return an error on SQL error", func() {
				mock.ExpectExec("INSERT INTO Groups").WillReturnError(fmt.Errorf("err"))

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

			repo = NewGroupRepo(db, zap.NewNop())
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

			repo = NewGroupRepo(db, zap.NewNop())
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

			g.It("Should return an empty array", func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Groups")).WillReturnRows(sqlmock.NewRows(cols))

				res, _ := repo.GetAll(context.TODO())

				Expect(res).To(Equal([]*domain.Group{}))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})
	})
}
