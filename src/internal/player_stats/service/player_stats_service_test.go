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
		var infractionRepo *mocks.InfractionRepo
		var service *pStatService
		var ctx context.Context

		g.BeforeEach(func() {
			infractionRepo = new(mocks.InfractionRepo)
			service = &pStatService{
				infractionRepo: infractionRepo,
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
	})
}
