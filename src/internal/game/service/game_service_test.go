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
	"fmt"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Game Service", func() {
		var service *gameService
		var gameRepo *mocks.GameRepo

		g.BeforeEach(func() {
			gameRepo = new(mocks.GameRepo)
			service = &gameService{
				repo:    gameRepo,
				timeout: time.Second * 2,
				games:   map[string]domain.Game{},
			}
		})

		g.Describe("AddGame()", func() {
			g.It("Should insert a new game into the games map", func() {
				mockGame := &mocks.Game{}
				mockGame.On("GetName").Return("mock")

				service.AddGame(mockGame)

				Expect(service.games[mockGame.GetName()]).ToNot(BeNil())
				Expect(service.games[mockGame.GetName()]).To(Equal(mockGame))
			})
		})

		g.Describe("GetAllGames()", func() {
			g.It("Should return all existing games", func() {
				mockGame1 := &mocks.Game{}
				mockGame1.On("GetName").Return("mock1")
				mockGame2 := &mocks.Game{}
				mockGame2.On("GetName").Return("mock2")
				mockGame3 := &mocks.Game{}
				mockGame3.On("GetName").Return("mock3")

				service.AddGame(mockGame1)
				service.AddGame(mockGame2)
				service.AddGame(mockGame3)

				allGames := service.GetAllGames()

				Expect(allGames).ToNot(BeNil())
				Expect(allGames).To(ContainElements(mockGame1, mockGame2, mockGame3))
			})
		})

		g.Describe("GameExists()", func() {
			g.It("Should return true if the game exists", func() {
				mockGame := &mocks.Game{}
				mockGame.On("GetName").Return("mock")

				service.AddGame(mockGame)

				Expect(service.GameExists(mockGame.GetName())).To(BeTrue())
			})

			g.It("Should return false if the game does not exist", func() {
				mockGame := &mocks.Game{}
				mockGame.On("GetName").Return("mock1")

				service.AddGame(mockGame)

				Expect(service.GameExists("nonexistent game")).To(BeFalse())
			})
		})

		g.Describe("GetGame()", func() {
			var mockGame *mocks.Game

			g.Describe("Success", func() {
				g.BeforeEach(func() {
					mockGame = &mocks.Game{}
					mockGame.On("GetName").Return("mock")

					service.AddGame(mockGame)
				})

				g.It("Should not return an error", func() {
					_, err := service.GetGame(mockGame.GetName())

					Expect(err).To(BeNil())
				})

				g.It("Should return a game which exists", func() {
					Expect(service.GetGame(mockGame.GetName())).To(Equal(mockGame))
				})
			})

			g.Describe("Fail", func() {
				g.BeforeEach(func() {
					mockGame = &mocks.Game{}
					mockGame.On("GetName").Return("mock")

					service.AddGame(mockGame)
				})

				g.It("Should return an error", func() {
					_, err := service.GetGame("nonexistent game")

					Expect(err).ToNot(BeNil())
				})

				g.It("Should return nil if the game does not exist", func() {
					game, _ := service.GetGame("nonexistent game")

					Expect(game).To(BeNil())
				})
			})
		})

		g.Describe("GetGameSettings()", func() {
			var game *mocks.Game

			g.BeforeEach(func() {
				game = new(mocks.Game)
				game.On("GetName").Return("testgame")
			})

			g.Describe("Success", func() {
				var expected *domain.GameSettings

				g.BeforeEach(func() {
					expected = &domain.GameSettings{
						BanCommandPattern: "test",
					}

					service.AddGame(game)
					gameRepo.On("GetSettings", mock.Anything).
						Return(expected, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.GetGameSettings(game.GetName())

					Expect(err).To(BeNil())
					gameRepo.AssertExpectations(t)
				})

				g.It("Should return the expected settings", func() {
					got, err := service.GetGameSettings(game.GetName())

					Expect(err).To(BeNil())
					Expect(got).To(Equal(expected))
					gameRepo.AssertExpectations(t)
				})
			})

			g.Describe("Invalid game provided", func() {
				g.It("Should return a domain.ErrNotFound error", func() {
					_, err := service.GetGameSettings(game.GetName())

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				})

				g.It("Should not call GetSettings repo method", func() {
					_, _ = service.GetGameSettings(game.GetName())

					gameRepo.AssertNotCalled(t, "GetSettings", mock.Anything)
				})
			})

			g.Describe("Repo error", func() {
				g.BeforeEach(func() {
					service.AddGame(game)
					gameRepo.On("GetSettings", mock.Anything).Return(nil, fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := service.GetGameSettings(game.GetName())

					Expect(err).ToNot(BeNil())
					gameRepo.AssertExpectations(t)
				})
			})
		})

		g.Describe("SetGameSettings()", func() {
			var game *mocks.Game

			g.BeforeEach(func() {
				game = new(mocks.Game)
				game.On("GetName").Return("testgame")
			})

			g.Describe("Success", func() {
				g.BeforeEach(func() {
					service.AddGame(game)
					gameRepo.On("SetSettings", mock.Anything, mock.Anything).Return(nil)
				})

				g.It("Should not return an error", func() {
					err := service.SetGameSettings(game.GetName(), &domain.GameSettings{})

					Expect(err).To(BeNil())
					gameRepo.AssertExpectations(t)
					game.AssertExpectations(t)
				})
			})

			g.Describe("Game not found", func() {
				g.It("Should return a domain.ErrNotFound error", func() {
					err := service.SetGameSettings("nonexistent", &domain.GameSettings{})

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
				})

				g.It("Should not call SetSettings on game repo", func() {
					_ = service.SetGameSettings("nonexistent", &domain.GameSettings{})

					gameRepo.AssertNotCalled(t, "SetSettings", mock.Anything, mock.Anything)
				})
			})

			g.Describe("Repo error", func() {
				g.BeforeEach(func() {
					service.AddGame(game)
					gameRepo.On("SetSettings", mock.Anything, mock.Anything).Return(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := service.SetGameSettings(game.GetName(), &domain.GameSettings{})

					Expect(err).ToNot(BeNil())
					gameRepo.AssertExpectations(t)
					game.AssertExpectations(t)
				})
			})
		})
	})
}
