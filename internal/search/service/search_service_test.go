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
	"github.com/guregu/null"
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

	g.Describe("Search service", func() {
		var service *searchService
		var playerRepo *mocks.PlayerRepo
		var infractionRepo *mocks.InfractionRepo
		var ctx context.Context

		g.BeforeEach(func() {
			playerRepo = new(mocks.PlayerRepo)
			infractionRepo = new(mocks.InfractionRepo)
			service = &searchService{
				playerRepo:     playerRepo,
				infractionRepo: infractionRepo,
				timeout:        time.Second * 2,
				logger:         zap.NewNop(),
			}
			ctx = context.TODO()
		})

		g.Describe("SearchPlayers()", func() {
			g.Describe("Successful search", func() {
				var results []*domain.Player

				g.BeforeEach(func() {
					results = []*domain.Player{
						{
							PlayerID:    "player1",
							Platform:    "platform",
							LastSeen:    time.Now(),
							CurrentName: "1-name",
						},
						{
							PlayerID:    "player2",
							Platform:    "platform",
							LastSeen:    time.Now(),
							CurrentName: "2-name",
						},
						{
							PlayerID:    "player3",
							Platform:    "platform",
							LastSeen:    time.Now(),
							CurrentName: "3-name",
						},
						{
							PlayerID:    "player4",
							Platform:    "platform",
							LastSeen:    time.Now(),
							CurrentName: "4-name",
						},
					}
				})

				g.Describe("Search type is 'name'", func() {
					g.BeforeEach(func() {
						playerRepo.On("SearchByName", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
							Return(len(results), results, nil)
					})

					g.It("Should not return an error", func() {
						_, _, err := service.SearchPlayers(ctx, "name", "name", "", 10, 0)

						Expect(err).To(BeNil())
						playerRepo.AssertExpectations(t)
					})

					g.It("Should return the correct results", func() {
						totalCount, results, err := service.SearchPlayers(ctx, "name", "name", "", 10, 0)

						Expect(err).To(BeNil())
						Expect(results).To(Equal(results))
						Expect(totalCount).To(Equal(len(results)))
						playerRepo.AssertExpectations(t)
					})
				})

				g.Describe("Search type is 'id'", func() {
					g.BeforeEach(func() {
						playerRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).
							Return(results[0], nil)
					})

					g.It("Should not return an error", func() {
						_, _, err := service.SearchPlayers(ctx, "playerid", "id", "platform", 10, 0)

						Expect(err).To(BeNil())
						playerRepo.AssertExpectations(t)
					})

					g.It("Should return the correct result", func() {
						totalCount, results, err := service.SearchPlayers(ctx, "playerid", "id", "platform", 10, 0)

						Expect(err).To(BeNil())
						Expect(results).To(Equal([]*domain.Player{results[0]}))
						Expect(totalCount).To(Equal(1))
						playerRepo.AssertExpectations(t)
					})
				})
			})
		})

		g.Describe("SearchInfractions()", func() {
			g.Describe("No valid search fields were provided", func() {
				g.It("Should return an error", func() {
					_, _, err := service.SearchInfractions(ctx, domain.FindArgs{"Invalid": "invalid"}, 0, 0)

					httpErr, ok := err.(*domain.HTTPError)

					Expect(ok).To(BeTrue())
					Expect(httpErr.Message).To(Equal("No search fields were provided"))
				})
			})

			g.Describe("Results found", func() {
				var results []*domain.Infraction

				g.BeforeEach(func() {
					results = []*domain.Infraction{
						{
							InfractionID: 1,
							PlayerID:     "playerid",
							Platform:     "platform",
							UserID:       null.NewString("userid", true),
							ServerID:     1,
							Type:         domain.InfractionTypeWarning,
							Reason:       null.NewString("reason", true),
							Duration:     null.Int{},
							SystemAction: true,
							CreatedAt:    null.Time{},
							ModifiedAt:   null.Time{},
							IssuerName:   "username",
						},
						{
							InfractionID: 2,
							PlayerID:     "playerid2",
							Platform:     "platform2",
							UserID:       null.NewString("userid", true),
							ServerID:     1,
							Type:         domain.InfractionTypeBan,
							Reason:       null.NewString("reason", true),
							Duration:     null.NewInt(60, true),
							SystemAction: false,
							CreatedAt:    null.Time{},
							ModifiedAt:   null.Time{},
							IssuerName:   "username",
						},
						{
							InfractionID: 1,
							PlayerID:     "playerid3",
							Platform:     "platform3",
							UserID:       null.NewString("userid", true),
							ServerID:     1,
							Type:         domain.InfractionTypeKick,
							Reason:       null.NewString("reason", true),
							Duration:     null.Int{},
							SystemAction: false,
							CreatedAt:    null.Time{},
							ModifiedAt:   null.Time{},
							IssuerName:   "username",
						},
					}

					infractionRepo.On("Search", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(100, results, nil)
				})

				g.It("Should not return an error", func() {
					_, _, err := service.SearchInfractions(ctx, domain.FindArgs{"UserID": "id"}, 0, 10)

					Expect(err).To(BeNil())
					mock.AssertExpectationsForObjects(t)
					infractionRepo.AssertExpectations(t)
				})

				g.It("Should return the correct results and total count", func() {
					total, got, err := service.SearchInfractions(ctx, domain.FindArgs{"UserID": "id"}, 0, 10)

					Expect(err).To(BeNil())
					Expect(got).To(Equal(results))
					Expect(total).To(Equal(100))
					mock.AssertExpectationsForObjects(t)
					infractionRepo.AssertExpectations(t)
				})
			})

			g.Describe("No results found, func()", func() {
				g.BeforeEach(func() {
					infractionRepo.On("Search", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(0, []*domain.Infraction{}, nil)
				})

				g.It("Should not return an error", func() {
					_, _, err := service.SearchInfractions(ctx, domain.FindArgs{"UserID": "id"}, 0, 10)

					Expect(err).To(BeNil())
					mock.AssertExpectationsForObjects(t)
					infractionRepo.AssertExpectations(t)
				})

				g.It("Should return an empty array and a total count of 0", func() {
					total, got, err := service.SearchInfractions(ctx, domain.FindArgs{"UserID": "id"}, 0, 10)

					Expect(err).To(BeNil())
					Expect(got).To(Equal([]*domain.Infraction{}))
					Expect(total).To(Equal(0))
					mock.AssertExpectationsForObjects(t)
					infractionRepo.AssertExpectations(t)
				})
			})
		})
	})
}
