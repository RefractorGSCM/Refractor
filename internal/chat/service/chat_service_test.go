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
		var websocketService *mocks.WebsocketService
		var service *chatService
		var ctx context.Context

		g.BeforeEach(func() {
			repo = new(mocks.ChatRepo)
			playerRepo = new(mocks.PlayerRepo)
			websocketService = new(mocks.WebsocketService)

			service = &chatService{
				repo:             repo,
				playerRepo:       playerRepo,
				websocketService: websocketService,
				timeout:          time.Second * 2,
				logger:           zap.NewNop(),
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
	})
}
