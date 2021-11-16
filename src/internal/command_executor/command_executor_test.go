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
	"context"
	"fmt"
	"github.com/franela/goblin"
	"github.com/guregu/null"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
)

// The command executor is one of the most heavily tested parts of Refractor. This is important because thee commands
// executed have a very real effect on player experience. We can't have the wrong commands being executed, or have
// commands which execute when they shouldn't, or it would be very detrimental to a game server.

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Command Executor", func() {
		var rconService *mocks.RCONService
		var gameService *mocks.GameService
		var serverRepo *mocks.ServerRepo
		var playerNameRepo *mocks.PlayerNameRepo
		var cmdexec *executor
		var game *mocks.Game
		var ctx context.Context

		g.BeforeEach(func() {
			rconService = new(mocks.RCONService)
			gameService = new(mocks.GameService)
			serverRepo = new(mocks.ServerRepo)
			playerNameRepo = new(mocks.PlayerNameRepo)
			cmdexec = &executor{
				rconService:    rconService,
				gameService:    gameService,
				serverRepo:     serverRepo,
				playerNameRepo: playerNameRepo,
				logger:         zap.NewNop(),
			}
			game = new(mocks.Game)
			ctx = context.TODO()
		})

		g.Describe("PrepareInfractionCommands()", func() {
			var infraction *domain.Infraction
			var serverID int64

			g.BeforeEach(func() {
				serverID = 6
				infraction = &domain.Infraction{
					InfractionID: 1,
					PlayerID:     "playerid1",
					Platform:     "platform1",
					ServerID:     serverID,
					Type:         domain.InfractionTypeBan,
					Reason:       null.NewString("test reason", true),
					Duration:     null.NewInt(420, true),
					PlayerName:   "Test Player Name",
				}
			})

			g.Describe("Successful prepare", func() {
				g.BeforeEach(func() {
					serverRepo.On("GetByID", mock.Anything, serverID).Return(&domain.Server{
						Game: "testgame",
					}, nil)
					gameService.On("GetGame", "testgame").Return(game, nil)
					gameService.On("GetGameSettings", mock.Anything).Return(&domain.GameSettings{
						Commands: &domain.GameCommandSettings{
							CreateInfractionCommands: &domain.InfractionCommands{
								Warn: []*domain.InfractionCommand{},
								Mute: []*domain.InfractionCommand{},
								Kick: []*domain.InfractionCommand{},
								Ban: []*domain.InfractionCommand{
									{
										Command:  "Ban {{PLAYER_NAME}} {{DURATION}} {{REASON}}",
										RunOnAll: true,
									},
								},
							},
							UpdateInfractionCommands: nil,
							DeleteInfractionCommands: nil,
							RepealInfractionCommands: &domain.InfractionCommands{
								Warn: []*domain.InfractionCommand{},
								Mute: []*domain.InfractionCommand{},
								Kick: []*domain.InfractionCommand{},
								Ban: []*domain.InfractionCommand{
									{
										Command:  "Test {{PLAYER_NAME}} {{PLAYER_ID}} {{PLATFORM}} {{DURATION}} {{REASON}}",
										RunOnAll: true,
									},
								},
							},
						},
					}, nil)
				})

				g.It("Should not return an error", func() {
					_, err := cmdexec.PrepareInfractionCommands(ctx, infraction, domain.InfractionCommandCreate, serverID)

					Expect(err).To(BeNil())
				})

				g.It("Should return a command payload with the correct values", func() {
					expectedCreateCommands := []*infractionCommand{
						{
							Command: fmt.Sprintf("Ban %s %d %s", infraction.PlayerName,
								infraction.Duration.ValueOrZero(), infraction.Reason.ValueOrZero()),
							RunOnAll: true,
							ServerID: serverID,
						},
					}
					expectedRepealCommands := []*infractionCommand{
						{
							Command: fmt.Sprintf("Test %s %s %s %d %s", infraction.PlayerName, infraction.PlayerID,
								infraction.Platform, infraction.Duration.ValueOrZero(), infraction.Reason.ValueOrZero()),
							RunOnAll: true,
							ServerID: serverID,
						},
					}

					createPayload, err := cmdexec.PrepareInfractionCommands(ctx, infraction, domain.InfractionCommandCreate, serverID)
					Expect(err).To(BeNil())
					for i, cmd := range createPayload.GetCommands() {
						Expect(cmd.GetServerID()).To(Equal(expectedCreateCommands[i].GetServerID()))
						Expect(cmd.GetCommand()).To(Equal(expectedCreateCommands[i].GetCommand()))
					}

					repealPayload, err := cmdexec.PrepareInfractionCommands(ctx, infraction, domain.InfractionCommandRepeal, serverID)
					Expect(err).To(BeNil())
					for i, cmd := range repealPayload.GetCommands() {
						Expect(cmd.GetServerID()).To(Equal(expectedRepealCommands[i].GetServerID()))
						Expect(cmd.GetCommand()).To(Equal(expectedRepealCommands[i].GetCommand()))
					}
				})

				g.Describe("Player was not set", func() {
					g.BeforeEach(func() {
						infraction.PlayerName = ""
						playerNameRepo.On("GetNames", mock.Anything, mock.Anything, mock.Anything).Return("name", nil, nil)
					})

					g.It("Should not return an error", func() {
						_, err := cmdexec.PrepareInfractionCommands(ctx, infraction, domain.InfractionCommandCreate, serverID)

						Expect(err).To(BeNil())
					})
				})
			})

			g.Describe("Server repo error", func() {
				g.BeforeEach(func() {
					serverRepo.On("GetByID", mock.Anything, serverID).Return(nil, fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := cmdexec.PrepareInfractionCommands(ctx, &domain.Infraction{PlayerName: "name"}, domain.InfractionCommandCreate, serverID)

					Expect(err).ToNot(BeNil())
					serverRepo.AssertExpectations(t)
				})
			})

			g.Describe("Game not found", func() {
				g.BeforeEach(func() {
					serverRepo.On("GetByID", mock.Anything, serverID).Return(&domain.Server{Game: "testgame"}, nil)
					gameService.On("GetGame", mock.Anything).Return(nil, domain.ErrNotFound)
				})

				g.It("Should return an error", func() {
					_, err := cmdexec.PrepareInfractionCommands(ctx, &domain.Infraction{PlayerName: "name"}, domain.InfractionCommandCreate, serverID)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					serverRepo.AssertExpectations(t)
					gameService.AssertExpectations(t)
				})
			})

			g.Describe("Game has no commands set", func() {
				g.BeforeEach(func() {
					serverRepo.On("GetByID", mock.Anything, serverID).Return(&domain.Server{Game: "testgame"}, nil)
					gameService.On("GetGame", mock.Anything).Return(game, nil)
					gameService.On("GetGameSettings", mock.Anything).Return(&domain.GameSettings{
						Commands: &domain.GameCommandSettings{},
					}, nil)
				})

				g.It("Should return an error", func() {
					_, err := cmdexec.PrepareInfractionCommands(ctx, &domain.Infraction{PlayerName: "name"}, domain.InfractionCommandCreate, serverID)

					Expect(err).ToNot(BeNil())
					serverRepo.AssertExpectations(t)
					gameService.AssertExpectations(t)
				})
			})

			g.Describe("Invalid action provided", func() {
				g.BeforeEach(func() {
					serverRepo.On("GetByID", mock.Anything, serverID).Return(&domain.Server{Game: "testgame"}, nil)
					gameService.On("GetGame", mock.Anything).Return(game, nil)
					gameService.On("GetGameSettings", mock.Anything).Return(&domain.GameSettings{
						Commands: &domain.GameCommandSettings{
							CreateInfractionCommands: &domain.InfractionCommands{},
							UpdateInfractionCommands: &domain.InfractionCommands{},
							DeleteInfractionCommands: &domain.InfractionCommands{},
							RepealInfractionCommands: &domain.InfractionCommands{},
						},
					}, nil)
				})

				g.It("Should return the correct error", func() {
					_, err := cmdexec.PrepareInfractionCommands(ctx, &domain.Infraction{PlayerName: "name"}, "invalid", serverID)

					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("no infraction action type: invalid"))
					serverRepo.AssertExpectations(t)
					gameService.AssertExpectations(t)
				})
			})

			g.Describe("Invalid infraction type provided", func() {
				g.BeforeEach(func() {
					serverRepo.On("GetByID", mock.Anything, serverID).Return(&domain.Server{Game: "testgame"}, nil)
					gameService.On("GetGame", mock.Anything).Return(game, nil)
					gameService.On("GetGameSettings", mock.Anything).Return(&domain.GameSettings{
						Commands: &domain.GameCommandSettings{
							CreateInfractionCommands: &domain.InfractionCommands{},
							UpdateInfractionCommands: &domain.InfractionCommands{},
							DeleteInfractionCommands: &domain.InfractionCommands{},
							RepealInfractionCommands: &domain.InfractionCommands{},
						},
					}, nil)
				})

				g.It("Should return the correct error", func() {
					_, err := cmdexec.PrepareInfractionCommands(ctx, &domain.Infraction{Type: "invalid", PlayerName: "name"},
						domain.InfractionCommandCreate, serverID)

					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("no commands found for infraction type: invalid"))
					serverRepo.AssertExpectations(t)
					gameService.AssertExpectations(t)
				})
			})
		})
	})
}
