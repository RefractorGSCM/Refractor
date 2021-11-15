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
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Player Stats Service", func() {
		var playerRepo *mocks.PlayerRepo
		var infractionRepo *mocks.InfractionRepo
		var gameService *mocks.GameService
		var service *pStatService
		var ctx context.Context

		g.BeforeEach(func() {
			playerRepo = new(mocks.PlayerRepo)
			infractionRepo = new(mocks.InfractionRepo)
			gameService = new(mocks.GameService)
			service = &pStatService{
				playerRepo:     playerRepo,
				infractionRepo: infractionRepo,
				gameService:    gameService,
				timeout:        time.Second * 2,
				logger:         zap.NewNop(),
			}
			ctx = context.TODO()
		})

		g.Describe("GetInfractionCount()", func() {
			g.Describe("Infraction count fetched", func() {
				g.BeforeEach(func() {
					infractionRepo.On("GetPlayerTotalInfractions", mock.Anything, "platform", "playerid").
						Return(1827, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.GetInfractionCount(ctx, "platform", "playerid")

					Expect(err).To(BeNil())
					infractionRepo.AssertExpectations(t)
				})

				g.It("Should return the correct infraction count", func() {
					count, err := service.GetInfractionCount(ctx, "platform", "playerid")

					Expect(err).To(BeNil())
					Expect(count).To(Equal(1827))
					infractionRepo.AssertExpectations(t)
				})
			})

			g.Describe("Player has no infractions", func() {
				g.BeforeEach(func() {
					infractionRepo.On("GetPlayerTotalInfractions", mock.Anything, "platform", "playerid").
						Return(0, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.GetInfractionCount(ctx, "platform", "playerid")

					Expect(err).To(BeNil())
					infractionRepo.AssertExpectations(t)
				})

				g.It("Should return 0 as the infraction count", func() {
					count, err := service.GetInfractionCount(ctx, "platform", "playerid")

					Expect(err).To(BeNil())
					Expect(count).To(Equal(0))
					infractionRepo.AssertExpectations(t)
				})
			})

			g.Describe("Repo error", func() {
				g.BeforeEach(func() {
					infractionRepo.On("GetPlayerTotalInfractions", mock.Anything, "platform", "playerid").
						Return(0, fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := service.GetInfractionCount(ctx, "platform", "playerid")

					Expect(err).ToNot(BeNil())
					infractionRepo.AssertExpectations(t)
				})
			})
		})

		g.Describe("GetPlayerPayload()", func() {
			var game *mocks.Game

			g.BeforeEach(func() {
				game = new(mocks.Game)
			})

			g.Describe("Success", func() {
				var expected *domain.PlayerPayload

				g.BeforeEach(func() {
					expected = &domain.PlayerPayload{
						Player: &domain.Player{
							PlayerID:      "playerid",
							Platform:      "platform",
							CurrentName:   "currentname",
							PreviousNames: []string{"previous"},
							Watched:       false,
							LastSeen:      time.Time{},
							CreatedAt:     time.Time{},
							ModifiedAt:    time.Time{},
						},
						InfractionCount:              16,
						InfractionCountSinceTimespan: 5,
					}

					gameService.On("GetGameSettings", mock.Anything).Return(&domain.GameSettings{
						General: &domain.GeneralSettings{
							PlayerInfractionTimespan: 1440, // 1 day in minutes
						},
					}, nil)

					playerRepo.On("GetByID", mock.Anything, "platform", "playerid").
						Return(expected.Player, nil)

					infractionRepo.On("GetPlayerTotalInfractions", mock.Anything, "platform", "playerid").
						Return(expected.InfractionCount, nil)
					infractionRepo.On("GetPlayerInfractionCountSince", mock.Anything, "platform", "playerid", mock.Anything).
						Return(expected.InfractionCountSinceTimespan, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.GetPlayerPayload(ctx, "platform", "playerid", game)

					Expect(err).To(BeNil())
					gameService.AssertExpectations(t)
					playerRepo.AssertExpectations(t)
					infractionRepo.AssertExpectations(t)
				})

				g.It("Should return the correct player payload", func() {
					payload, err := service.GetPlayerPayload(ctx, "platform", "playerid", game)

					Expect(err).To(BeNil())
					Expect(payload).To(Equal(expected))
					gameService.AssertExpectations(t)
					playerRepo.AssertExpectations(t)
					infractionRepo.AssertExpectations(t)
				})
			})

			g.Describe("Game service error", func() {
				g.BeforeEach(func() {
					gameService.On("GetGameSettings", mock.Anything).Return(nil, fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := service.GetPlayerPayload(ctx, "platform", "playerid", game)

					Expect(err).ToNot(BeNil())
					gameService.AssertExpectations(t)
				})
			})

			g.Describe("Player repo error", func() {
				g.BeforeEach(func() {
					gameService.On("GetGameSettings", mock.Anything).Return(&domain.GameSettings{
						General: &domain.GeneralSettings{
							PlayerInfractionTimespan: 1440, // 1 day in minutes
						},
					}, nil)

					playerRepo.On("GetByID", mock.Anything, "platform", "playerid").
						Return(nil, fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := service.GetPlayerPayload(ctx, "platform", "playerid", game)

					Expect(err).ToNot(BeNil())
					gameService.AssertExpectations(t)
					playerRepo.AssertExpectations(t)
				})
			})

			g.Describe("Infraction repo error", func() {
				g.BeforeEach(func() {
					gameService.On("GetGameSettings", mock.Anything).Return(&domain.GameSettings{
						General: &domain.GeneralSettings{
							PlayerInfractionTimespan: 1440, // 1 day in minutes
						},
					}, nil)

					playerRepo.On("GetByID", mock.Anything, "platform", "playerid").
						Return(&domain.Player{
							PlayerID: "playerid",
							Platform: "platform",
						}, nil)
				})

				g.Describe("GetPlayerTotalInfractions error", func() {
					g.BeforeEach(func() {
						infractionRepo.On("GetPlayerTotalInfractions", mock.Anything, "platform", "playerid").
							Return(0, fmt.Errorf("err"))
					})

					g.It("Should return an error", func() {
						_, err := service.GetPlayerPayload(ctx, "platform", "playerid", game)

						Expect(err).ToNot(BeNil())
						gameService.AssertExpectations(t)
						playerRepo.AssertExpectations(t)
						infractionRepo.AssertExpectations(t)
					})
				})

				g.Describe("GetInfractionCountSince error", func() {
					g.BeforeEach(func() {
						infractionRepo.On("GetPlayerTotalInfractions", mock.Anything, "platform", "playerid").
							Return(10, nil)
						infractionRepo.On("GetPlayerInfractionCountSince", mock.Anything, "platform", "playerid", mock.Anything).
							Return(0, fmt.Errorf("err"))
					})

					g.It("Should return an error", func() {
						_, err := service.GetPlayerPayload(ctx, "platform", "playerid", game)

						Expect(err).ToNot(BeNil())
						gameService.AssertExpectations(t)
						playerRepo.AssertExpectations(t)
						infractionRepo.AssertExpectations(t)
					})
				})
			})
		})
	})
}
