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
		var playerRepo *mocks.PlayerRepo
		var serverRepo *mocks.ServerRepo
		var service *infractionService
		var ctx = context.TODO()

		g.BeforeEach(func() {
			mockRepo = new(mocks.InfractionRepo)
			playerRepo = new(mocks.PlayerRepo)
			serverRepo = new(mocks.ServerRepo)
			service = &infractionService{
				repo:            mockRepo,
				playerRepo:      playerRepo,
				serverRepo:      serverRepo,
				timeout:         time.Second * 2,
				logger:          zap.NewNop(),
				infractionTypes: getInfractionTypes(),
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
					playerRepo.On("Exists", mock.Anything, mock.Anything).Return(true, nil)
					serverRepo.On("Exists", mock.Anything, mock.Anything).Return(true, nil)
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

			g.Describe("Player not found", func() {
				g.BeforeEach(func() {
					playerRepo.On("Exists", mock.Anything, mock.Anything).Return(false, nil)
				})

				g.It("Should return an HTTP error", func() {
					_, err := service.Store(ctx, &domain.Infraction{
						Platform: "platform",
						PlayerID: "playerid",
					})

					Expect(err).ToNot(BeNil())

					httpErr, ok := err.(*domain.HTTPError)

					Expect(ok).To(BeTrue())
					Expect(httpErr.Message).To(Equal("Player not found"))
					mockRepo.AssertExpectations(t)
					playerRepo.AssertExpectations(t)
				})
			})

			g.Describe("Server not found", func() {
				g.BeforeEach(func() {
					playerRepo.On("Exists", mock.Anything, mock.Anything).Return(true, nil)
					serverRepo.On("Exists", mock.Anything, mock.Anything).Return(false, nil)
				})

				g.It("Should return an HTTP error", func() {
					_, err := service.Store(ctx, &domain.Infraction{
						Platform: "platform",
						PlayerID: "playerid",
					})

					Expect(err).ToNot(BeNil())

					httpErr, ok := err.(*domain.HTTPError)

					Expect(ok).To(BeTrue())
					Expect(httpErr.Message).To(Equal("Server not found"))
					mockRepo.AssertExpectations(t)
					playerRepo.AssertExpectations(t)
				})
			})

			g.Describe("Repository error", func() {
				g.BeforeEach(func() {
					playerRepo.On("Exists", mock.Anything, mock.Anything).Return(true, nil)
					serverRepo.On("Exists", mock.Anything, mock.Anything).Return(true, nil)
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

		g.Describe("filterUpdateArgs()", func() {
			var infraction *domain.Infraction
			var args domain.UpdateArgs

			g.BeforeEach(func() {
				infraction = &domain.Infraction{}

				args = domain.UpdateArgs{
					// we include both reason and duration for all types to see if they will filter the update args correctly
					"Reason":   "Updated Reason",
					"Duration": null.NewInt(1000, true),
					"UserID":   null.NewString("updatedid", true), // not an allowed update field, should be ignore
				}
			})

			g.Describe("Warning", func() {
				g.BeforeEach(func() {
					infraction.Type = domain.InfractionTypeWarning
				})

				g.It("Should not return an error", func() {
					_, err := service.filterUpdateArgs(ctx, infraction, args)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should only return args with allowed fields", func() {
					expected := domain.UpdateArgs{
						"Reason": "Updated Reason",
					}

					args, err := service.filterUpdateArgs(ctx, infraction, args)

					Expect(err).To(BeNil())
					Expect(args).To(Equal(expected))
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Mute", func() {
				g.BeforeEach(func() {
					infraction.Type = domain.InfractionTypeMute
				})

				g.It("Should not return an error", func() {
					_, err := service.filterUpdateArgs(ctx, infraction, args)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should only return args with allowed fields", func() {
					expected := domain.UpdateArgs{
						"Reason":   "Updated Reason",
						"Duration": null.NewInt(1000, true),
					}

					args, err := service.filterUpdateArgs(ctx, infraction, args)

					Expect(err).To(BeNil())
					Expect(args).To(Equal(expected))
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Kick", func() {
				g.BeforeEach(func() {
					infraction.Type = domain.InfractionTypeKick
				})

				g.It("Should not return an error", func() {
					_, err := service.filterUpdateArgs(ctx, infraction, args)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should only return args with allowed fields", func() {
					expected := domain.UpdateArgs{
						"Reason": "Updated Reason",
					}

					args, err := service.filterUpdateArgs(ctx, infraction, args)

					Expect(err).To(BeNil())
					Expect(args).To(Equal(expected))
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Ban", func() {
				g.BeforeEach(func() {
					infraction.Type = domain.InfractionTypeBan
				})

				g.It("Should not return an error", func() {
					_, err := service.filterUpdateArgs(ctx, infraction, args)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should only return args with allowed fields", func() {
					expected := domain.UpdateArgs{
						"Reason":   "Updated Reason",
						"Duration": null.NewInt(1000, true),
					}

					args, err := service.filterUpdateArgs(ctx, infraction, args)

					Expect(err).To(BeNil())
					Expect(args).To(Equal(expected))
					mockRepo.AssertExpectations(t)
				})
			})
		})
	})
}
