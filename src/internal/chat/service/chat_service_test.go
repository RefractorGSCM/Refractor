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
	kratos "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Chat Service", func() {
		var repo *mocks.ChatRepo
		var playerRepo *mocks.PlayerRepo
		var playerNameRepo *mocks.PlayerNameRepo
		var serverService *mocks.ServerService
		var websocketService *mocks.WebsocketService
		var flaggedWordService *mocks.FlaggedWordService
		var authorizer *mocks.Authorizer
		var service *chatService
		var ctx context.Context

		g.BeforeEach(func() {
			repo = new(mocks.ChatRepo)
			playerRepo = new(mocks.PlayerRepo)
			playerNameRepo = new(mocks.PlayerNameRepo)
			serverService = new(mocks.ServerService)
			websocketService = new(mocks.WebsocketService)
			flaggedWordService = new(mocks.FlaggedWordService)
			authorizer = new(mocks.Authorizer)

			service = &chatService{
				repo:               repo,
				playerRepo:         playerRepo,
				playerNameRepo:     playerNameRepo,
				serverService:      serverService,
				websocketService:   websocketService,
				flaggedWordService: flaggedWordService,
				authorizer:         authorizer,
				timeout:            time.Second * 2,
				logger:             zap.NewNop(),
			}

			ctx = context.TODO()
		})

		g.Describe("Store()", func() {
			var message *domain.ChatMessage

			g.BeforeEach(func() {
				message = &domain.ChatMessage{
					PlayerID: "playerid",
					Platform: "platform",
					ServerID: 1,
					Message:  "test message",
					Flagged:  false,
				}
			})

			g.Describe("Successful store", func() {
				g.BeforeEach(func() {
					flaggedWordService.On("MessageContainsFlaggedWord", mock.Anything, mock.Anything).Return(false, nil)
					repo.On("Store", mock.Anything, mock.Anything).Return(nil)
				})

				g.It("Should not return an error", func() {
					err := service.Store(ctx, message)

					Expect(err).To(BeNil())
					repo.AssertExpectations(t)
				})
			})

			g.Describe("Repo error", func() {
				g.BeforeEach(func() {
					flaggedWordService.On("MessageContainsFlaggedWord", mock.Anything, mock.Anything).Return(false, nil)
					repo.On("Store", mock.Anything, mock.Anything).Return(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := service.Store(ctx, message)

					Expect(err).ToNot(BeNil())
					repo.AssertExpectations(t)
				})
			})
		})

		g.Describe("HandleChatReceive()", func() {
			var zapCore zapcore.Core
			var recordedLogs *observer.ObservedLogs
			var body *domain.ChatReceiveBody

			g.BeforeEach(func() {
				// Since HandleChatReceive does not return error, we can check if any error occurred by the logger output.
				// To do this, we use zap's built in observer library to watch for Error messages.
				zapCore, recordedLogs = observer.New(zapcore.ErrorLevel)
				service.logger = zap.New(zapCore)
				body = &domain.ChatReceiveBody{
					ServerID:   1,
					PlayerID:   "playerid",
					Platform:   "platform",
					Name:       "playername",
					Message:    "test chat message",
					SentByUser: false,
				}
			})

			g.Describe("Successful message broadcast and storage", func() {
				g.BeforeEach(func() {
					websocketService.On("BroadcastServerMessage", mock.Anything, mock.Anything, mock.Anything).
						Return(nil)
					playerRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(&domain.Player{
						PlayerID:    body.PlayerID,
						Platform:    body.Platform,
						CurrentName: body.Name,
					}, nil)
					flaggedWordService.On("MessageContainsFlaggedWord", mock.Anything, mock.Anything).Return(false, nil)
					repo.On("Store", mock.Anything, mock.Anything).Return(nil)
				})

				g.It("Should not log any errors", func() {
					service.HandleChatReceive(body, body.ServerID, nil)

					Expect(recordedLogs.All()).To(Equal([]observer.LoggedEntry{}))
					websocketService.AssertExpectations(t)
					playerRepo.AssertExpectations(t)
					repo.AssertExpectations(t)
				})
			})

			g.Describe("Websocket broadcast error", func() {
				g.BeforeEach(func() {
					zapCore, recordedLogs = observer.New(zapcore.WarnLevel)
					service.logger = zap.New(zapCore)

					websocketService.On("BroadcastServerMessage", mock.Anything, mock.Anything, mock.Anything).
						Return(fmt.Errorf("broadcast error"))
					playerRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(&domain.Player{
						PlayerID:    body.PlayerID,
						Platform:    body.Platform,
						CurrentName: body.Name,
					}, nil)
					repo.On("Store", mock.Anything, mock.Anything).Return(nil)
					flaggedWordService.On("MessageContainsFlaggedWord", mock.Anything, mock.Anything).Return(false, nil)
				})

				g.It("Should only log one error of level Warning", func() {
					service.HandleChatReceive(body, body.ServerID, nil)

					Expect(len(recordedLogs.All())).To(Equal(1))
					repo.AssertExpectations(t)
					playerRepo.AssertExpectations(t)
					repo.AssertExpectations(t)
				})
			})

			g.Describe("Player repo error", func() {
				g.BeforeEach(func() {
					websocketService.On("BroadcastServerMessage", mock.Anything, mock.Anything, mock.Anything).
						Return(nil)
					playerRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(&domain.Player{
						PlayerID:    body.PlayerID,
						Platform:    body.Platform,
						CurrentName: body.Name,
					}, fmt.Errorf("player repo err"))
				})

				g.It("Should only log one error of level Error", func() {
					service.HandleChatReceive(body, body.ServerID, nil)

					Expect(len(recordedLogs.All())).To(Equal(1))
					repo.AssertExpectations(t)
					playerRepo.AssertExpectations(t)

					// Since the function should return in case of a player repo error, we can verify this by making sure
					// that it never reached the point of storing the chat message.
					repo.AssertNotCalled(t, "Store")
				})
			})

			g.Describe("Chat repo error", func() {
				g.BeforeEach(func() {
					websocketService.On("BroadcastServerMessage", mock.Anything, mock.Anything, mock.Anything).
						Return(nil)
					playerRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(&domain.Player{
						PlayerID:    body.PlayerID,
						Platform:    body.Platform,
						CurrentName: body.Name,
					}, nil)
					repo.On("Store", mock.Anything, mock.Anything).Return(fmt.Errorf("repo err"))
					flaggedWordService.On("MessageContainsFlaggedWord", mock.Anything, mock.Anything).Return(false, nil)
				})

				g.It("Should only log one error of level Error", func() {
					service.HandleChatReceive(body, body.ServerID, nil)

					Expect(len(recordedLogs.All())).To(Equal(1))
					repo.AssertExpectations(t)
					playerRepo.AssertExpectations(t)
					repo.AssertExpectations(t)
				})
			})
		})

		g.Describe("GetRecentByServer()", func() {
			g.Describe("Results found", func() {
				var messages []*domain.ChatMessage

				g.BeforeEach(func() {
					messages = []*domain.ChatMessage{
						{
							MessageID: 1,
							PlayerID:  "playerid",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 1",
							Flagged:   false,
							Name:      "playername",
						}, {
							MessageID: 2,
							PlayerID:  "playerid2",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 2",
							Flagged:   false,
							Name:      "playername",
						}, {
							MessageID: 3,
							PlayerID:  "playerid3",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 3",
							Flagged:   false,
							Name:      "playername",
						}, {
							MessageID: 4,
							PlayerID:  "playerid4",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 4",
							Flagged:   false,
							Name:      "playername",
						}, {
							MessageID: 5,
							PlayerID:  "playerid5",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 5",
							Flagged:   false,
							Name:      "playername",
						}, {
							MessageID: 6,
							PlayerID:  "playerid6",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 6",
							Flagged:   false,
							Name:      "playername",
						}, {
							MessageID: 7,
							PlayerID:  "playerid7",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 7",
							Flagged:   false,
							Name:      "playername",
						}, {
							MessageID: 8,
							PlayerID:  "playerid8",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 8",
							Flagged:   false,
							Name:      "playername",
						}, {
							MessageID: 9,
							PlayerID:  "playerid9",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 9",
							Flagged:   false,
							Name:      "playername",
						}, {
							MessageID: 10,
							PlayerID:  "playerid10",
							Platform:  "platform",
							ServerID:  1,
							Message:   "message 10",
							Flagged:   false,
							Name:      "playername",
						},
					}

					repo.On("GetRecentByServer", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("int")).
						Return(messages, nil)
					playerNameRepo.On("GetNames", mock.Anything, mock.Anything, mock.Anything).
						Return("playername", nil, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.GetRecentByServer(ctx, 1, 10)

					Expect(err).To(BeNil())
					repo.AssertExpectations(t)
				})

				g.It("Should return the correct results", func() {
					results, err := service.GetRecentByServer(ctx, 1, 10)

					Expect(err).To(BeNil())
					Expect(results).To(Equal(messages))
					repo.AssertExpectations(t)
				})
			})

			g.Describe("No results found", func() {
				g.BeforeEach(func() {
					repo.On("GetRecentByServer", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("int")).
						Return(nil, domain.ErrNotFound)
				})

				g.It("Should not return an error, and should return an empty slice", func() {
					got, err := service.GetRecentByServer(ctx, 1, 10)

					Expect(err).To(BeNil())
					Expect(got).To(Equal([]*domain.ChatMessage{}))
					repo.AssertExpectations(t)
				})
			})
		})

		g.Describe("GetFlaggedMessages()", func() {
			var servers []*domain.Server
			var messages []*domain.ChatMessage

			g.BeforeEach(func() {
				servers = []*domain.Server{
					{
						ID: 1,
					},
					{
						ID: 2,
					},
					{
						ID: 3,
					},
				}

				messages = []*domain.ChatMessage{
					{
						MessageID: 1,
						PlayerID:  "playerid1",
						Platform:  "platform",
						ServerID:  1,
						Message:   "message 1",
						Flagged:   true,
					},
					{
						MessageID: 2,
						PlayerID:  "playerid2",
						Platform:  "platform",
						ServerID:  1,
						Message:   "message 2",
						Flagged:   true,
					},
					{
						MessageID: 3,
						PlayerID:  "playerid3",
						Platform:  "platform",
						ServerID:  3,
						Message:   "message 3",
						Flagged:   true,
					},
				}
			})

			g.Describe("User was not provided in context", func() {
				g.AfterEach(func() {
					authorizer.AssertNotCalled(t, "HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				})

				g.Describe("Results found", func() {
					g.BeforeEach(func() {
						serverService.On("GetAll", mock.Anything).Return(servers, nil)
						repo.On("GetFlaggedMessages", mock.Anything, mock.Anything, mock.Anything).Return(messages, nil)
						playerNameRepo.On("GetNames", mock.Anything, mock.Anything, mock.Anything).Return("currentName", []string{}, nil)
					})

					g.It("Should not return an error", func() {
						_, err := service.GetFlaggedMessages(ctx, 20)

						Expect(err).To(BeNil())
						serverService.AssertExpectations(t)
						repo.AssertExpectations(t)
					})

					g.It("Should return the correct results", func() {
						got, err := service.GetFlaggedMessages(ctx, 20)

						Expect(err).To(BeNil())
						Expect(got).To(Equal(messages))
						serverService.AssertExpectations(t)
						repo.AssertExpectations(t)
					})
				})

				g.Describe("No servers found", func() {
					g.BeforeEach(func() {
						serverService.On("GetAll", mock.Anything).Return(nil, domain.ErrNotFound)
					})

					g.It("Should not return an error", func() {
						_, err := service.GetFlaggedMessages(ctx, 20)

						Expect(err).To(BeNil())
						serverService.AssertExpectations(t)
					})

					g.It("Should return an empty slice", func() {
						got, err := service.GetFlaggedMessages(ctx, 20)

						Expect(err).To(BeNil())
						Expect(got).To(Equal([]*domain.ChatMessage{}))
						serverService.AssertExpectations(t)
					})
				})

				g.Describe("No results found", func() {
					g.BeforeEach(func() {
						serverService.On("GetAll", mock.Anything).Return(servers, nil)
						repo.On("GetFlaggedMessages", mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrNotFound)
					})

					g.It("Should not return an error", func() {
						_, err := service.GetFlaggedMessages(ctx, 20)

						Expect(err).To(BeNil())
						serverService.AssertExpectations(t)
					})

					g.It("Should return an empty slice", func() {
						got, err := service.GetFlaggedMessages(ctx, 20)

						Expect(err).To(BeNil())
						Expect(got).To(Equal([]*domain.ChatMessage{}))
						serverService.AssertExpectations(t)
					})
				})
			})

			g.Describe("User was provided in context", func() {
				g.BeforeEach(func() {
					au := &domain.AuthUser{
						Session: &kratos.Session{
							Identity: kratos.Identity{
								Id: "testuserid",
							},
						},
					}

					ctx = context.WithValue(ctx, "user", au)
				})

				g.Describe("Results were found", func() {
					g.BeforeEach(func() {
						messages = append([]*domain.ChatMessage{}, messages[0], messages[2])

						serverService.On("GetAll", mock.Anything).Return(servers, nil)
						repo.On("GetFlaggedMessages", mock.Anything, mock.Anything, mock.Anything).Return(messages, nil)
						authorizer.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
							Return(true, nil).Once() // server ID 1
						authorizer.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
							Return(false, nil).Once() // server ID 2
						authorizer.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
							Return(true, nil).Once() // server ID 3
						playerNameRepo.On("GetNames", mock.Anything, mock.Anything, mock.Anything).Return("currentName", []string{}, nil)
					})

					g.It("Should not return an error", func() {
						_, err := service.GetFlaggedMessages(ctx, 20)

						Expect(err).To(BeNil())
						serverService.AssertExpectations(t)
						authorizer.AssertExpectations(t)
						repo.AssertExpectations(t)
					})

					g.It("Should return the correct values", func() {
						got, err := service.GetFlaggedMessages(ctx, 20)

						Expect(err).To(BeNil())
						Expect(got).To(Equal(messages))
						serverService.AssertExpectations(t)
						authorizer.AssertExpectations(t)
						repo.AssertExpectations(t)
					})
				})
			})
		})
	})
}
