package service

import (
	"Refractor/domain"
	"Refractor/domain/mocks"
	"context"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
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
			g.It("Should not return an error", func() {
				mockRepo.On("GetAll", mock.Anything).Return([]*domain.Group{}, nil)

				_, err := service.GetAll(ctx)

				Expect(err).To(BeNil())
			})

			g.It("Should return the correct groups", func() {
				mockRepo.On("GetAll", mock.Anything).Return(mockGroups, nil)

				foundGroups, err := service.GetAll(ctx)

				Expect(err).To(BeNil())
				Expect(foundGroups).To(Equal(mockGroups))
			})
		})
	})
}
