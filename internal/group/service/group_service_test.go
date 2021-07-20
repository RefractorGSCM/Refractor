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

package service

import (
	"Refractor/domain"
	"Refractor/domain/mocks"
	"context"
	"fmt"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"math"
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	var mockRepo *mocks.GroupRepo
	var service domain.GroupService
	var ctx = context.TODO()

	g.Describe("Store()", func() {
		g.BeforeEach(func() {
			mockRepo = new(mocks.GroupRepo)
			service = NewGroupService(mockRepo, time.Second*2)
		})

		g.Describe("Group stored successfully", func() {
			g.It("Should not return an error", func() {
				mockRepo.On("Store", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)

				err := service.Store(ctx, &domain.Group{Name: "Test Group"})

				Expect(err).To(BeNil())
				mockRepo.AssertExpectations(t)
			})
		})
	})

	g.Describe("GetByID()", func() {
		g.BeforeEach(func() {
			mockRepo = new(mocks.GroupRepo)
			service = NewGroupService(mockRepo, time.Second*2)
		})

		g.Describe("Result fetched successfully", func() {
			g.It("Should not return an error", func() {
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("int64")).Return(&domain.Group{}, nil)

				_, err := service.GetByID(ctx, 1)

				Expect(err).To(BeNil())
			})

			g.It("Should return the correct group", func() {
				mockGroup := &domain.Group{
					ID:          1,
					Name:        "Test Group",
					Color:       5423552,
					Position:    15,
					Permissions: "345276874377",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				}

				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("int64")).Return(mockGroup, nil)

				foundGroup, err := service.GetByID(ctx, 1)

				Expect(err).To(BeNil())
				Expect(foundGroup).To(Equal(mockGroup))
			})
		})
	})

	g.Describe("GetAll()", func() {
		var mockGroups []*domain.Group

		g.BeforeEach(func() {
			mockRepo = new(mocks.GroupRepo)
			service = NewGroupService(mockRepo, time.Second*2)
			mockGroups = []*domain.Group{}

			mockGroups = append(mockGroups, &domain.Group{
				ID:          1,
				Name:        "Test Group",
				Color:       5423552,
				Position:    15,
				Permissions: "345276874377",
				CreatedAt:   time.Time{},
				ModifiedAt:  time.Time{},
			})

			mockGroups = append(mockGroups, &domain.Group{
				ID:          2,
				Name:        "Test Group 2",
				Color:       542355452,
				Position:    14,
				Permissions: "34527324326874377",
				CreatedAt:   time.Time{},
				ModifiedAt:  time.Time{},
			})

			mockGroups = append(mockGroups, &domain.Group{
				ID:          3,
				Name:        "Test Group 3",
				Color:       452355452,
				Position:    6,
				Permissions: "44554645664534434",
				CreatedAt:   time.Time{},
				ModifiedAt:  time.Time{},
			})
		})

		g.Describe("Results fetched successfully", func() {
			var baseGroup *domain.Group

			g.BeforeEach(func() {
				baseGroup = &domain.Group{
					ID:          -1,
					Name:        "Everyone",
					Color:       0xcecece,
					Position:    math.MaxInt32,
					Permissions: "2738628437",
				}

				mockRepo.On("GetBaseGroup", mock.Anything).Return(baseGroup, nil)
			})

			g.It("Should not return an error", func() {
				mockRepo.On("GetAll", mock.Anything).Return([]*domain.Group{}, nil)

				_, err := service.GetAll(ctx)

				Expect(err).To(BeNil())
			})

			g.It("Should return the correct groups", func() {
				mockRepo.On("GetAll", mock.Anything).Return(mockGroups, nil)

				foundGroups, err := service.GetAll(ctx)

				expected := mockGroups
				expected = append(expected, baseGroup)

				Expect(err).To(BeNil())
				Expect(foundGroups).To(Equal(expected))
			})

			g.It("Should return the groups sorted ascendingly by position", func() {
				mockRepo.On("GetAll", mock.Anything).Return(mockGroups, nil)

				expected := mockGroups
				expected = append(mockGroups, baseGroup)
				expected = domain.GroupSlice(expected).SortByPosition()

				foundGroups, err := service.GetAll(ctx)

				Expect(err).To(BeNil())
				Expect(foundGroups).To(Equal(expected))
			})
		})
	})

	g.Describe("Delete()", func() {
		g.BeforeEach(func() {
			mockRepo = new(mocks.GroupRepo)
			service = NewGroupService(mockRepo, time.Second*2)
		})

		g.Describe("Target group found", func() {
			g.BeforeEach(func() {
				mockRepo.On("Delete", mock.Anything, mock.AnythingOfType("int64")).Return(nil)
			})

			g.It("Should not return an error", func() {
				err := service.Delete(context.TODO(), 1)

				Expect(err).To(BeNil())
				mockRepo.AssertExpectations(t)
			})
		})

		g.Describe("Target group was not found", func() {
			g.It("Should return the domain.ErrNotFound error", func() {
				mockRepo.On("Delete", mock.Anything, mock.AnythingOfType("int64")).Return(domain.ErrNotFound)

				err := service.Delete(context.TODO(), 1)

				Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				mockRepo.AssertExpectations(t)
			})
		})
	})

	g.Describe("Update()", func() {
		var updatedGroup *domain.Group

		g.BeforeEach(func() {
			mockRepo = new(mocks.GroupRepo)
			service = NewGroupService(mockRepo, time.Second*2)

			updatedGroup = &domain.Group{
				ID:          1,
				Name:        "Updated Group",
				Color:       0xcecece,
				Position:    6,
				Permissions: "7456223",
				CreatedAt:   time.Time{},
				ModifiedAt:  time.Time{},
			}
		})

		g.Describe("Target group found", func() {
			g.BeforeEach(func() {
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("int64"),
					mock.AnythingOfType("domain.UpdateArgs")).Return(updatedGroup, nil)
			})

			g.It("Should not return an error", func() {
				_, err := service.Update(context.TODO(), 1, domain.UpdateArgs{})

				Expect(err).To(BeNil())
				mockRepo.AssertExpectations(t)
			})

			g.It("Should return the updated group", func() {
				updated, err := service.Update(context.TODO(), 1, domain.UpdateArgs{})

				Expect(err).To(BeNil())
				Expect(updated).To(Equal(updatedGroup))
				mockRepo.AssertExpectations(t)
			})
		})

		g.Describe("Target group was not found", func() {
			g.BeforeEach(func() {
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("int64"),
					mock.AnythingOfType("domain.UpdateArgs")).Return(nil, domain.ErrNotFound)
			})

			g.It("Should return the domain.ErrNotFound error", func() {
				_, err := service.Update(context.TODO(), 1, domain.UpdateArgs{})

				Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				mockRepo.AssertExpectations(t)
			})

			g.It("Should return nil as the group", func() {
				g, _ := service.Update(context.TODO(), 1, domain.UpdateArgs{})

				Expect(g).To(BeNil())
				mockRepo.AssertExpectations(t)
			})
		})
	})

	g.Describe("Reorder()", func() {
		g.BeforeEach(func() {
			mockRepo = new(mocks.GroupRepo)
			service = NewGroupService(mockRepo, time.Second*2)
		})

		g.Describe("Reorder success", func() {
			g.BeforeEach(func() {
				mockRepo.On("Reorder", mock.Anything, mock.AnythingOfType("[]*domain.GroupReorderInfo")).
					Return(nil)
			})

			g.It("Should not return an error", func() {
				err := service.Reorder(context.TODO(), []*domain.GroupReorderInfo{})

				Expect(err).To(BeNil())
				mockRepo.AssertExpectations(t)
			})
		})

		g.Describe("Reorder error", func() {
			g.BeforeEach(func() {
				mockRepo.On("Reorder", mock.Anything, mock.AnythingOfType("[]*domain.GroupReorderInfo")).
					Return(fmt.Errorf(""))
			})

			g.It("Should return an error", func() {
				err := service.Reorder(context.TODO(), []*domain.GroupReorderInfo{})

				Expect(err).ToNot(BeNil())
				mockRepo.AssertExpectations(t)
			})
		})
	})

	g.Describe("UpdateBase()", func() {
		var baseGroup *domain.Group
		var updateArgs domain.UpdateArgs

		g.BeforeEach(func() {
			mockRepo = new(mocks.GroupRepo)
			service = NewGroupService(mockRepo, time.Second*2)

			baseGroup = &domain.Group{
				ID:          -1,
				Name:        "Everyone",
				Color:       0xcecece,
				Position:    math.MaxInt32,
				Permissions: "1",
				CreatedAt:   time.Time{},
				ModifiedAt:  time.Time{},
			}

			updateArgs = domain.UpdateArgs{
				"Color":       0xececec,
				"Permissions": "2",
			}
		})

		g.Describe("Successful update", func() {
			g.BeforeEach(func() {
				mockRepo.On("GetBaseGroup", mock.Anything).Return(baseGroup, nil)
				mockRepo.On("SetBaseGroup", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)
			})

			g.It("Should not return an error", func() {
				_, err := service.UpdateBase(context.TODO(), updateArgs)

				Expect(err).To(BeNil())
				mock.AssertExpectationsForObjects(t)
			})

			g.It("Should return the updated group", func() {
				expected := &domain.Group{
					ID:          -1,
					Name:        "Everyone",
					Color:       updateArgs["Color"].(int),
					Position:    math.MaxInt32,
					Permissions: updateArgs["Permissions"].(string),
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				}

				updated, _ := service.UpdateBase(context.TODO(), updateArgs)

				Expect(updated).To(Equal(expected))
				mock.AssertExpectationsForObjects(t)
			})
		})

		g.Describe("Repository error", func() {
			g.It("Should return an error on GetBaseGroup repo error", func() {
				mockRepo.On("GetBaseGroup", mock.Anything).Return(nil, fmt.Errorf("getbasegroup error"))

				_, err := service.UpdateBase(context.TODO(), updateArgs)

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("getbasegroup error"))
			})

			g.It("Should return an error on SetBaseGroup repo error", func() {
				mockRepo.On("GetBaseGroup", mock.Anything).Return(baseGroup, nil)
				mockRepo.On("SetBaseGroup", mock.Anything, mock.AnythingOfType("*domain.Group")).
					Return(fmt.Errorf("setbasegroup error"))

				_, err := service.UpdateBase(context.TODO(), updateArgs)

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("setbasegroup error"))
			})
		})
	})
}
