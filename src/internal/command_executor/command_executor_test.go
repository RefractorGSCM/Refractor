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

package command_executor

import (
	"Refractor/domain"
	"Refractor/domain/mocks"
	"fmt"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Command Executor", func() {
		var rconService *mocks.RCONService
		var rconClient *mocks.RCONClient
		var gameService *mocks.GameService
		var cmdexec *executor
		var payload *domain.PlayerCommandPayload
		var game *mocks.Game

		g.BeforeEach(func() {
			rconService = new(mocks.RCONService)
			rconClient = new(mocks.RCONClient)
			gameService = new(mocks.GameService)
			cmdexec = &executor{
				rconService: rconService,
				gameService: gameService,
				logger:      zap.NewNop(),
			}
			payload = &domain.PlayerCommandPayload{
				PlayerID: "id",
				Platform: "platform",
				Name:     "name",
				Duration: 1234,
				Reason:   "test reason",
			}
			game = new(mocks.Game)
		})

		g.Describe("RunInfractionCommands()", func() {
			g.Describe("Successful execute", func() {
				g.BeforeEach(func() {
					gameService.On("GetGameSettings", mock.Anything).Return(&domain.GameSettings{
						Commands: &domain.GameCommandSettings{
							CreateInfractionCommands: &domain.InfractionCommands{
								Warn: []string{},
								Mute: []string{},
								Kick: []string{"test {{PLAYER_NAME}} {{PLAYER_ID}} {{DURATION}} {{REASON}}"},
								Ban:  []string{"ban {{PLAYER_NAME}} {{REASON}}"},
							},
							UpdateInfractionCommands: &domain.InfractionCommands{
								Warn: []string{},
								Mute: []string{},
								Kick: []string{},
								Ban:  []string{},
							},
							DeleteInfractionCommands: &domain.InfractionCommands{
								Warn: []string{},
								Mute: []string{},
								Kick: []string{},
								Ban:  []string{"unban {{PLAYER_NAME}}"},
							},
							RepealInfractionCommands: &domain.InfractionCommands{
								Warn: []string{},
								Mute: []string{},
								Kick: []string{},
								Ban:  []string{"unban {{PLAYER_NAME}}"},
							},
						},
					}, nil)

					expectedKickCommand := fmt.Sprintf("test %s %s %d %s", payload.Name, payload.PlayerID,
						payload.Duration, payload.Reason)
					expectedBanCommand := fmt.Sprintf("ban %s %s", payload.Name, payload.Reason)

					rconService.On("GetServerClient", mock.Anything).Return(rconClient, nil)

					rconClient.On("ExecCommand", expectedKickCommand).Return("res", nil).Once()
					rconClient.On("ExecCommand", expectedBanCommand).Return("res", nil).Once()
				})

				g.It("Should not return an error", func() {
					err := cmdexec.RunInfractionCommands(domain.InfractionTypeKick, domain.InfractionCommandCreate, payload, 1, game)
					Expect(err).To(BeNil())

					err = cmdexec.RunInfractionCommands(domain.InfractionTypeBan, domain.InfractionCommandCreate, payload, 1, game)
					Expect(err).To(BeNil())

					rconClient.AssertExpectations(t)
				})
			})

			g.Describe("Game not found", func() {
				g.BeforeEach(func() {
					gameService.On("GetGameSettings", mock.Anything).Return(nil, domain.ErrNotFound)

					game.On("GetName").Return("non-existent")
				})

				g.It("Should return an error", func() {
					err := cmdexec.RunInfractionCommands(domain.InfractionTypeKick, domain.InfractionCommandCreate, payload, 1, game)

					Expect(err).ToNot(BeNil())
				})

				g.It("Should not run any commands", func() {
					rconService.AssertNotCalled(t, "ExecCommand", mock.Anything)
					rconService.AssertExpectations(t)
				})
			})

			g.Describe("Invalid infraction type", func() {
				g.BeforeEach(func() {
					gameService.On("GetGameSettings", mock.Anything).Return(&domain.GameSettings{Commands: &domain.GameCommandSettings{}}, nil)
				})

				g.It("Should return an error", func() {
					err := cmdexec.RunInfractionCommands("invalid", domain.InfractionCommandCreate, payload, 1, game)

					Expect(err).ToNot(BeNil())
					gameService.AssertExpectations(t)
				})

				g.It("Should not run any commands", func() {
					rconService.AssertNotCalled(t, "ExecCommand", mock.Anything)
					rconService.AssertExpectations(t)
				})
			})
		})
	})
}
