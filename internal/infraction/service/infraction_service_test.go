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
	"github.com/guregu/null"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Infraction Service", func() {
		var mockRepo *mocks.InfractionRepo
		var service *infractionService
		var ctx = context.TODO()

		g.BeforeEach(func() {
			mockRepo = new(mocks.InfractionRepo)
			service = &infractionService{
				repo:    mockRepo,
				timeout: time.Second * 2,
				logger:  zap.NewNop(),
			}
		})

		g.Describe("Store()", func() {
			g.Describe("Infraction stored successfully", func() {
				var mockInfraction *domain.Infraction

				g.BeforeEach(func() {
					mockInfraction = &domain.Infraction{
						InfractionID: 1,
						PlayerID:     "playerid",
						Platform:     "platform",
						UserID:       null.NewString("userid", true),
						ServerID:     4,
						Type:         domain.InfractionTypeKick,
						Reason:       null.NewString("Test reason", true),
						Duration:     null.Int{},
						SystemAction: false,
						CreatedAt:    null.Time{},
						ModifiedAt:   null.Time{},
					}

					mockRepo.On("Store", mock.Anything, mock.Anything).Return(mockInfraction, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.Store(ctx, mockInfraction)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should return the correct infraction", func() {
					infraction, err := service.Store(ctx, mockInfraction)

					Expect(err).To(BeNil())
					Expect(infraction).To(Equal(mockInfraction))
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Repository error", func() {
				g.BeforeEach(func() {
					mockRepo.On("Store", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := service.Store(ctx, &domain.Infraction{})

					Expect(err).ToNot(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})
		})

		g.Describe("GetByID()", func() {
			g.Describe("Result found", func() {
				var mockInfraction *domain.Infraction

				g.BeforeEach(func() {
					mockInfraction = &domain.Infraction{
						InfractionID: 1,
						PlayerID:     "playerid",
						Platform:     "platform",
						UserID:       null.NewString("userid", true),
						ServerID:     4,
						Type:         domain.InfractionTypeWarning,
						Reason:       null.NewString("Test reason", true),
						Duration:     null.Int{},
						SystemAction: false,
						CreatedAt:    null.Time{},
						ModifiedAt:   null.Time{},
					}

					mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(mockInfraction, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.GetByID(ctx, 1)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should return the correct infraction", func() {
					foundInfraction, err := service.GetByID(ctx, 1)

					Expect(err).To(BeNil())
					Expect(foundInfraction).To(Equal(mockInfraction))
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("No results found", func() {
				g.BeforeEach(func() {
					mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, domain.ErrNotFound)
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					_, err := service.GetByID(ctx, 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					mockRepo.AssertExpectations(t)
				})
			})
		})
	})
}
