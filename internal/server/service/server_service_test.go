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

	var mockRepo *mocks.ServerRepo
	var service domain.ServerService
	var ctx = context.TODO()

	g.Describe("Store()", func() {
		g.BeforeEach(func() {
			mockRepo = new(mocks.ServerRepo)
			service = NewServerService(mockRepo, time.Second*2)
		})

		g.Describe("Server stored successfully", func() {
			g.It("Should not return an error", func() {
				mockRepo.On("Store", mock.Anything, mock.AnythingOfType("*domain.Server")).Return(nil)

				err := service.Store(ctx, &domain.Server{Name: "Test Server"})

				Expect(err).To(BeNil())
				mockRepo.AssertExpectations(t)
			})
		})
	})

	g.Describe("GetByID()", func() {
		g.BeforeEach(func() {
			mockRepo = new(mocks.ServerRepo)
			service = NewServerService(mockRepo, time.Second*2)
		})

		g.Describe("Result fetched successfully", func() {
			g.It("Should not return an error", func() {
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("int64")).Return(&domain.Server{}, nil)

				_, err := service.GetByID(ctx, 1)

				Expect(err).To(BeNil())
			})

			g.It("Should return the correct server", func() {
				mockServer := &domain.Server{
					ID:           1,
					Game:         "Test Game",
					Name:         "Test Server",
					Address:      "127.0.0.1",
					RCONPort:     4372,
					RCONPassword: "sjghjuwfxgdwfhij",
					CreatedAt:    time.Time{},
					ModifiedAt:   time.Time{},
				}

				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("int64")).Return(mockServer, nil)

				foundServer, err := service.GetByID(ctx, 1)

				Expect(err).To(BeNil())
				Expect(foundServer).To(Equal(mockServer))
			})
		})
	})
}
