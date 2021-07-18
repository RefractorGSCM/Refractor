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
}
