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
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"regexp"
	"sync"
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("RCON Service", func() {
		var service *rconService
		var gameService *mocks.GameService
		var game *mocks.Game
		var gameConfig *domain.GameConfig
		var serverID int64 = 1
		var rconClient *mocks.RCONClient
		var mockServer *domain.Server
		var clientCreator *mocks.ClientCreator

		g.BeforeEach(func() {
			rconClient = new(mocks.RCONClient)
			gameService = new(mocks.GameService)
			clientCreator = new(mocks.ClientCreator)

			service = &rconService{
				logger: zap.NewNop(),
				clients: map[int64]domain.RCONClient{
					serverID: rconClient,
				},
				gameService:   gameService,
				clientCreator: clientCreator,
				prevPlayers:   map[int64]map[string]*domain.OnlinePlayer{},
			}

			mockServer = &domain.Server{
				ID:           1,
				Game:         "TestGame",
				Name:         "Test Server",
				Address:      "127.0.0.1",
				RCONPort:     "1234",
				RCONPassword: "RconPassword",
				Deactivated:  false,
				CreatedAt:    time.Time{},
				ModifiedAt:   time.Time{},
			}

			// Game setup
			game = new(mocks.Game)
			gameConfig = &domain.GameConfig{
				UseRCON:                   true,
				AlivePingInterval:         time.Second * 1,
				EnableBroadcasts:          true,
				BroadcastPatterns:         map[string]*regexp.Regexp{},
				IgnoredBroadcastPatterns:  []*regexp.Regexp{},
				EnableChat:                false,
				PlayerListPollingInterval: time.Second * 1,
			}

			game.On("GetName").Return("TestGame")
			game.On("GetConfig").Return(gameConfig)
			game.On("GetPlayerListCommand").Return("PlayerList")
			game.On("GetCommandOutputPatterns").Return(&domain.CommandOutputPatterns{
				PlayerList: regexp.MustCompile("(?P<PlayerID>[0-9]+), (?P<Name>[a-zA-z0-9]+)"),
			})

			gameService.On("GetGame", mock.Anything).Return(game, nil)
		})

		g.Describe("CreateClient()", func() {
			g.Describe("Success", func() {
				g.BeforeEach(func() {
					gameService.On("GameExists", mock.AnythingOfType("string")).Return(true)
					clientCreator.On("GetClientFromConfig", mock.Anything, mock.Anything).Return(rconClient, nil)
					rconClient.On("SetBroadcastHandler", mock.Anything)
					rconClient.On("SetDisconnectHandler", mock.Anything)
					rconClient.On("Connect").Return(nil)
					rconClient.On("ListenForBroadcasts", mock.Anything, mock.Anything)
					rconClient.On("ExecCommand", mock.Anything).Return("", nil)
					rconClient.On("Close").Return(nil)
					rconClient.On("WaitGroup").Return(&sync.WaitGroup{})
				})

				g.It("Should not return an error", func() {
					err := service.CreateClient(mockServer)

					Expect(err).To(BeNil())
				})
			})

			g.Describe("ClientCreator error", func() {
				g.BeforeEach(func() {
					gameService.On("GameExists", mock.AnythingOfType("string")).Return(true)
					clientCreator.On("GetClientFromConfig", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err"))
					rconClient.On("Close").Return(nil)
					rconClient.On("WaitGroup").Return(&sync.WaitGroup{})
				})

				g.It("Should return an error", func() {
					err := service.CreateClient(mockServer)

					Expect(err).ToNot(BeNil())
				})
			})

			g.Describe("Client Connect error", func() {
				g.BeforeEach(func() {
					gameService.On("GameExists", mock.AnythingOfType("string")).Return(true)
					clientCreator.On("GetClientFromConfig", mock.Anything, mock.Anything).Return(rconClient, nil)
					rconClient.On("SetBroadcastHandler", mock.Anything)
					rconClient.On("SetDisconnectHandler", mock.Anything)
					rconClient.On("Connect").Return(fmt.Errorf("err"))
					rconClient.On("Close").Return(nil)
					rconClient.On("WaitGroup").Return(&sync.WaitGroup{})
				})

				g.It("Should return an error", func() {
					err := service.CreateClient(mockServer)

					Expect(err).ToNot(BeNil())
				})
			})
		})

		g.Describe("getOnlinePlayers()", func() {
			g.Describe("Success", func() {
				var rawOutput string
				var expected []*domain.OnlinePlayer

				g.BeforeEach(func() {
					rawOutput = `1, Player1
								2, Player2
								3, Player3
								4, Player4
								5, Player5`

					expected = []*domain.OnlinePlayer{
						{
							PlayerID: "1",
							Name:     "Player1",
						}, {
							PlayerID: "2",
							Name:     "Player2",
						}, {
							PlayerID: "3",
							Name:     "Player3",
						}, {
							PlayerID: "4",
							Name:     "Player4",
						}, {
							PlayerID: "5",
							Name:     "Player5",
						},
					}

					rconClient.On("ExecCommand", mock.Anything).Return(rawOutput, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.getOnlinePlayers(serverID, game)

					Expect(err).To(BeNil())
				})

				g.It("Should return a correct slice of online player data", func() {
					onlinePlayers, _ := service.getOnlinePlayers(serverID, game)

					Expect(onlinePlayers).To(Equal(expected))
				})
			})

			g.Describe("ExecCommand error", func() {
				g.BeforeEach(func() {
					rconClient.On("ExecCommand", mock.Anything).Return("", fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := service.getOnlinePlayers(serverID, game)

					Expect(err).ToNot(BeNil())
				})
			})
		})
	})
}
