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
	"Refractor/pkg/bitperms"
	"Refractor/pkg/perms"
	"context"
	"fmt"
	"github.com/franela/goblin"
	"github.com/guregu/null"
	. "github.com/onsi/gomega"
	kratos "github.com/ory/kratos-client-go"
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
		var playerNameRepo *mocks.PlayerNameRepo
		var serverRepo *mocks.ServerRepo
		var userMetaRepo *mocks.UserMetaRepo
		var authorizer *mocks.Authorizer
		var service *infractionService
		var ctx = context.TODO()

		g.BeforeEach(func() {
			mockRepo = new(mocks.InfractionRepo)
			playerRepo = new(mocks.PlayerRepo)
			playerNameRepo = new(mocks.PlayerNameRepo)
			serverRepo = new(mocks.ServerRepo)
			userMetaRepo = new(mocks.UserMetaRepo)
			authorizer = new(mocks.Authorizer)
			service = &infractionService{
				repo:            mockRepo,
				playerRepo:      playerRepo,
				playerNameRepo:  playerNameRepo,
				serverRepo:      serverRepo,
				userMetaRepo:    userMetaRepo,
				authorizer:      authorizer,
				timeout:         time.Second * 2,
				logger:          zap.NewNop(),
				infractionTypes: getInfractionTypes(),
			}
			ctx = context.TODO()
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
					_, err := service.Store(ctx, mockInfraction, nil, nil)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should return the correct infraction", func() {
					infraction, err := service.Store(ctx, mockInfraction, nil, nil)

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
					}, nil, nil)

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
					}, nil, nil)

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
					_, err := service.Store(ctx, &domain.Infraction{}, nil, nil)

					Expect(err).ToNot(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})
		})

		g.Describe("GetByID()", func() {
			g.Describe("User was not provided in context (do not check auth)", func() {
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
						userMetaRepo.On("GetUsername", mock.Anything, mock.Anything).Return("username", nil)
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

			g.Describe("User was set in context (check authorization)", func() {
				g.BeforeEach(func() {
					ctx = context.WithValue(ctx, "user", &domain.AuthUser{
						Session: &kratos.Session{
							Identity: kratos.Identity{
								Id: "userid",
							},
						},
					})

					mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domain.Infraction{
						ServerID: 1,
					}, nil)

					userMetaRepo.On("GetUsername", mock.Anything, mock.Anything).Return("username", nil)
				})

				g.Describe("User has authorization", func() {
					g.BeforeEach(func() {
						authorizer.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
							Return(true, nil)
					})

					g.It("Should not return an error", func() {
						_, err := service.GetByID(ctx, 1)

						Expect(err).To(BeNil())
						mockRepo.AssertExpectations(t)
					})
				})

				g.Describe("User does not have authorization", func() {
					g.BeforeEach(func() {
						authorizer.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
							Return(false, nil)
					})

					g.It("Should return an error", func() {
						_, err := service.GetByID(ctx, 1)

						Expect(err).ToNot(BeNil())
						mockRepo.AssertExpectations(t)
					})
				})
			})
		})

		g.Describe("GetByPlayer()", func() {
			var mockInfractions []*domain.Infraction

			g.BeforeEach(func() {
				mockInfractions = []*domain.Infraction{
					{
						InfractionID: 1,
						PlayerID:     "playerid",
						Platform:     "platform",
						UserID:       null.NewString("userid", true),
						ServerID:     1,
						Type:         domain.InfractionTypeMute,
						Reason:       null.NewString("Test mute reason", true),
						Duration:     null.NewInt(60, true),
						SystemAction: false,
						CreatedAt:    null.NewTime(time.Now(), true),
						ModifiedAt:   null.Time{},
						IssuerName:   "foundIssuerName",
					},
					{
						InfractionID: 2,
						PlayerID:     "playerid",
						Platform:     "platform",
						UserID:       null.NewString("userid", true),
						ServerID:     2,
						Type:         domain.InfractionTypeKick,
						Reason:       null.NewString("Test kick reason", true),
						Duration:     null.NewInt(0, false),
						SystemAction: false,
						CreatedAt:    null.NewTime(time.Now(), true),
						ModifiedAt:   null.Time{},
						IssuerName:   "foundIssuerName",
					},
					{
						InfractionID: 3,
						PlayerID:     "playerid",
						Platform:     "platform",
						UserID:       null.NewString("userid", true),
						ServerID:     3,
						Type:         domain.InfractionTypeWarning,
						Reason:       null.NewString("Test warn reason", true),
						Duration:     null.NewInt(0, false),
						SystemAction: false,
						CreatedAt:    null.NewTime(time.Now(), true),
						ModifiedAt:   null.Time{},
						IssuerName:   "foundIssuerName",
					},
					{
						InfractionID: 4,
						PlayerID:     "playerid",
						Platform:     "platform",
						UserID:       null.NewString("userid", true),
						ServerID:     4,
						Type:         domain.InfractionTypeBan,
						Reason:       null.NewString("Test ban reason", true),
						Duration:     null.NewInt(1440, true),
						SystemAction: false,
						CreatedAt:    null.NewTime(time.Now(), true),
						ModifiedAt:   null.Time{},
						IssuerName:   "foundIssuerName",
					},
				}
			})

			g.Describe("User was provided in context (check auth)", func() {
				g.BeforeEach(func() {
					ctx = context.WithValue(ctx, "user", &domain.AuthUser{
						Session: &kratos.Session{
							Identity: kratos.Identity{
								Id: "userid",
							},
						},
					})
				})

				g.Describe("Infractions were found", func() {
					g.BeforeEach(func() {
						mockRepo.On("GetByPlayer", mock.Anything, mock.Anything, mock.Anything).Return(mockInfractions, nil)
						serverRepo.On("GetAll", mock.Anything).Return([]*domain.Server{
							{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4},
						}, nil)

						// user has permission for servers ID 1 and 4 and is denied permission for servers ID 2 and 3. Notice the order of calls.
						authorizer.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
							Return(true, nil).Once() // ID 1
						authorizer.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
							Return(false, nil).Once() // ID 2
						authorizer.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
							Return(false, nil).Once() // ID 3
						authorizer.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
							Return(true, nil).Once() // ID 4

						userMetaRepo.On("GetUsername", mock.Anything, mock.Anything).Return("foundIssuerUsername", nil)
					})

					g.It("Should not return an error", func() {
						_, err := service.GetByPlayer(ctx, "playerid", "platform")

						Expect(err).To(BeNil())
						mockRepo.AssertExpectations(t)
					})

					g.It("Should return the correct infractions", func() {
						var expected []*domain.Infraction
						expected = append(expected, mockInfractions[0])
						expected = append(expected, mockInfractions[3])

						got, err := service.GetByPlayer(ctx, "playerid", "platform")

						Expect(err).To(BeNil())
						Expect(got).To(Equal(expected))
						mockRepo.AssertExpectations(t)
					})
				})

				g.Describe("Infractions not found", func() {
					g.BeforeEach(func() {
						mockRepo.On("GetByPlayer", mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrNotFound)
					})

					g.It("Should not return an error", func() {
						_, err := service.GetByPlayer(ctx, "playerid", "platform")

						Expect(err).To(BeNil())
						mockRepo.AssertExpectations(t)
					})

					g.It("Should return an empty array", func() {
						output, err := service.GetByPlayer(ctx, "playerid", "platform")

						Expect(err).To(BeNil())
						Expect(output).To(Equal([]*domain.Infraction{}))
						mockRepo.AssertExpectations(t)
					})
				})
			})

			g.Describe("User was not provided in context (don't check auth)", func() {
				g.Describe("Infractions were found", func() {
					g.BeforeEach(func() {
						mockRepo.On("GetByPlayer", mock.Anything, mock.Anything, mock.Anything).Return(mockInfractions, nil)
						userMetaRepo.On("GetUsername", mock.Anything, mock.Anything).Return("foundIssuerUsername", nil)
					})

					g.It("Should not return an error", func() {
						_, err := service.GetByPlayer(ctx, "playerid", "platform")

						Expect(err).To(BeNil())
						mockRepo.AssertExpectations(t)
					})

					g.It("Should return the correct infractions", func() {
						got, err := service.GetByPlayer(ctx, "playerid", "platform")

						Expect(err).To(BeNil())
						Expect(got).To(Equal(mockInfractions))
						mockRepo.AssertExpectations(t)
					})
				})

				g.Describe("No results were found", func() {
					g.BeforeEach(func() {
						mockRepo.On("GetByPlayer", mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrNotFound)
					})

					g.It("Should not return an error", func() {
						_, err := service.GetByPlayer(ctx, "playerid", "platform")

						Expect(err).To(BeNil())
						mockRepo.AssertExpectations(t)
					})

					g.It("Should return an empty array", func() {
						output, err := service.GetByPlayer(ctx, "playerid", "platform")

						Expect(err).To(BeNil())
						Expect(output).To(Equal([]*domain.Infraction{}))
						mockRepo.AssertExpectations(t)
					})
				})
			})
		})

		g.Describe("hasUpdatePermissions()", func() {
			var infraction *domain.Infraction
			var user *domain.AuthUser

			g.BeforeEach(func() {
				infraction = &domain.Infraction{
					InfractionID: 1,
					PlayerID:     "playerid",
					Platform:     "platform",
					UserID:       null.NewString("userid", true),
					ServerID:     1,
					Type:         domain.InfractionTypeWarning,
					Reason:       null.NewString("reason", true),
					Duration:     null.Int{},
					SystemAction: false,
					CreatedAt:    null.Time{},
					ModifiedAt:   null.Time{},
				}

				user = &domain.AuthUser{
					Session: &kratos.Session{
						Identity: kratos.Identity{
							Id: "anotheruserid",
						},
					},
				}
			})

			g.Describe("User is an admin", func() {
				g.BeforeEach(func() {
					adminPerms := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagAdministrator)).GetPermission()
					authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(adminPerms, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.hasUpdatePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					authorizer.AssertExpectations(t)
				})

				g.It("Should return true", func() {
					hasPermission, err := service.hasUpdatePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					Expect(hasPermission).To(BeTrue())
					authorizer.AssertExpectations(t)
				})
			})

			g.Describe("User is super admin", func() {
				g.BeforeEach(func() {
					superPerms := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagSuperAdmin)).GetPermission()
					authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(superPerms, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.hasUpdatePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					authorizer.AssertExpectations(t)
				})

				g.It("Should return true", func() {
					hasPermission, err := service.hasUpdatePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					Expect(hasPermission).To(BeTrue())
					authorizer.AssertExpectations(t)
				})
			})

			g.Describe("User created the infraction and they have permission to edit infractions they created", func() {
				g.BeforeEach(func() {
					infraction.UserID = null.NewString(user.Identity.Id, true)
					permissions := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagEditOwnInfractions)).GetPermission()
					authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(permissions, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.hasUpdatePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					authorizer.AssertExpectations(t)
				})

				g.It("Should return true", func() {
					hasPermission, err := service.hasUpdatePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					Expect(hasPermission).To(BeTrue())
					authorizer.AssertExpectations(t)
				})
			})

			g.Describe("User did not create the infraction but they have permission to edit any infraction", func() {
				g.BeforeEach(func() {
					infraction.UserID = null.NewString("not the right userid", true)
					permissions := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagEditAnyInfractions)).GetPermission()
					authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(permissions, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.hasUpdatePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					authorizer.AssertExpectations(t)
				})

				g.It("Should return true", func() {
					hasPermission, err := service.hasUpdatePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					Expect(hasPermission).To(BeTrue())
					authorizer.AssertExpectations(t)
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
					_, err := service.filterUpdateArgs(infraction, args)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should only return args with allowed fields", func() {
					expected := domain.UpdateArgs{
						"Reason": "Updated Reason",
					}

					args, err := service.filterUpdateArgs(infraction, args)

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
					_, err := service.filterUpdateArgs(infraction, args)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should only return args with allowed fields", func() {
					expected := domain.UpdateArgs{
						"Reason":   "Updated Reason",
						"Duration": null.NewInt(1000, true),
					}

					args, err := service.filterUpdateArgs(infraction, args)

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
					_, err := service.filterUpdateArgs(infraction, args)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should only return args with allowed fields", func() {
					expected := domain.UpdateArgs{
						"Reason": "Updated Reason",
					}

					args, err := service.filterUpdateArgs(infraction, args)

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
					_, err := service.filterUpdateArgs(infraction, args)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should only return args with allowed fields", func() {
					expected := domain.UpdateArgs{
						"Reason":   "Updated Reason",
						"Duration": null.NewInt(1000, true),
					}

					args, err := service.filterUpdateArgs(infraction, args)

					Expect(err).To(BeNil())
					Expect(args).To(Equal(expected))
					mockRepo.AssertExpectations(t)
				})
			})
		})

		g.Describe("hasDeletePermissions()", func() {
			var infraction *domain.Infraction
			var user *domain.AuthUser

			g.BeforeEach(func() {
				infraction = &domain.Infraction{
					InfractionID: 1,
					PlayerID:     "playerid",
					Platform:     "platform",
					UserID:       null.NewString("userid", true),
					ServerID:     1,
					Type:         domain.InfractionTypeWarning,
					Reason:       null.NewString("reason", true),
					Duration:     null.Int{},
					SystemAction: false,
					CreatedAt:    null.Time{},
					ModifiedAt:   null.Time{},
				}

				user = &domain.AuthUser{
					Session: &kratos.Session{
						Identity: kratos.Identity{
							Id: "anotheruserid",
						},
					},
				}
			})

			g.Describe("User is an admin", func() {
				g.BeforeEach(func() {
					adminPerms := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagAdministrator)).GetPermission()
					authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(adminPerms, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.hasDeletePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					authorizer.AssertExpectations(t)
				})

				g.It("Should return true", func() {
					hasPermission, err := service.hasDeletePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					Expect(hasPermission).To(BeTrue())
					authorizer.AssertExpectations(t)
				})
			})

			g.Describe("User is super admin", func() {
				g.BeforeEach(func() {
					superPerms := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagSuperAdmin)).GetPermission()
					authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(superPerms, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.hasDeletePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					authorizer.AssertExpectations(t)
				})

				g.It("Should return true", func() {
					hasPermission, err := service.hasDeletePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					Expect(hasPermission).To(BeTrue())
					authorizer.AssertExpectations(t)
				})
			})

			g.Describe("User created the infraction and they have permission to edit infractions they created", func() {
				g.BeforeEach(func() {
					infraction.UserID = null.NewString(user.Identity.Id, true)
					permissions := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagDeleteOwnInfractions)).GetPermission()
					authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(permissions, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.hasDeletePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					authorizer.AssertExpectations(t)
				})

				g.It("Should return true", func() {
					hasPermission, err := service.hasDeletePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					Expect(hasPermission).To(BeTrue())
					authorizer.AssertExpectations(t)
				})
			})

			g.Describe("User did not create the infraction but they have permission to edit any infraction", func() {
				g.BeforeEach(func() {
					infraction.UserID = null.NewString("not the right userid", true)
					permissions := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagDeleteAnyInfractions)).GetPermission()
					authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(permissions, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.hasDeletePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					authorizer.AssertExpectations(t)
				})

				g.It("Should return true", func() {
					hasPermission, err := service.hasDeletePermissions(ctx, infraction, user)

					Expect(err).To(BeNil())
					Expect(hasPermission).To(BeTrue())
					authorizer.AssertExpectations(t)
				})
			})
		})

		g.Describe("GetLinkedChatMessages()", func() {
			var expected []*domain.ChatMessage

			g.Describe("Results found", func() {
				g.BeforeEach(func() {
					expected = []*domain.ChatMessage{
						{
							MessageID: 1,
							PlayerID:  "playerid1",
							Platform:  "platform",
							ServerID:  1,
							Message:   "msg1",
							Flagged:   true,
						},
						{
							MessageID: 2,
							PlayerID:  "playerid2",
							Platform:  "platform",
							ServerID:  1,
							Message:   "msg2",
							Flagged:   true,
						},
						{
							MessageID: 3,
							PlayerID:  "playerid3",
							Platform:  "platform",
							ServerID:  1,
							Message:   "msg33",
							Flagged:   true,
						},
					}

					mockRepo.On("GetLinkedChatMessages", mock.Anything, mock.AnythingOfType("int64")).
						Return(expected, nil)
					playerNameRepo.On("GetNames", mock.Anything, mock.Anything, mock.Anything).
						Return("name", []string{}, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.GetLinkedChatMessages(ctx, 1)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should return the expected results", func() {
					got, err := service.GetLinkedChatMessages(ctx, 1)

					Expect(err).To(BeNil())
					Expect(got).To(Equal(expected))
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("No results not found", func() {
				g.BeforeEach(func() {
					expected = []*domain.ChatMessage{}

					mockRepo.On("GetLinkedChatMessages", mock.Anything, mock.AnythingOfType("int64")).
						Return(expected, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.GetLinkedChatMessages(ctx, 1)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should return an empty slice", func() {
					got, err := service.GetLinkedChatMessages(ctx, 1)

					Expect(err).To(BeNil())
					Expect(got).To(Equal(expected))
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Repo error", func() {
				g.BeforeEach(func() {
					mockRepo.On("GetLinkedChatMessages", mock.Anything, mock.AnythingOfType("int64")).
						Return(nil, fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := service.GetLinkedChatMessages(ctx, 1)

					Expect(err).ToNot(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})
		})

		g.Describe("LinkChatMessages()", func() {
			g.Describe("Successful link", func() {
				g.BeforeEach(func() {
					mockRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Infraction{}, nil)
					mockRepo.On("LinkChatMessages", mock.Anything, int64(1), int64(2)).Return(nil)
				})

				g.It("Should not return an error", func() {
					err := service.LinkChatMessages(ctx, 1, 2)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Infraction not found", func() {
				g.BeforeEach(func() {
					mockRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, domain.ErrNotFound)
				})

				g.It("Should return an HTTPError domain.ErrNotFound error", func() {
					err := service.LinkChatMessages(ctx, 1, 10)

					_, ok := err.(*domain.HTTPError)

					Expect(ok).To(BeTrue())
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Repository error", func() {
				g.BeforeEach(func() {
					mockRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Infraction{}, nil)
					mockRepo.On("LinkChatMessages", mock.Anything, int64(1), int64(10)).Return(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := service.LinkChatMessages(ctx, 1, 10)

					Expect(err).ToNot(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})
		})

		g.Describe("UnlinkChatMessages()", func() {
			g.Describe("Successful unlink", func() {
				g.BeforeEach(func() {
					mockRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Infraction{}, nil)
					mockRepo.On("UnlinkChatMessages", mock.Anything, int64(1), int64(2)).Return(nil)
				})

				g.It("Should not return an error", func() {
					err := service.UnlinkChatMessages(ctx, 1, 2)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Infraction not found", func() {
				g.BeforeEach(func() {
					mockRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, domain.ErrNotFound)
				})

				g.It("Should return an HTTPError domain.ErrNotFound error", func() {
					err := service.UnlinkChatMessages(ctx, 1, 10)

					_, ok := err.(*domain.HTTPError)

					Expect(ok).To(BeTrue())
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Repository error", func() {
				g.BeforeEach(func() {
					mockRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Infraction{}, nil)
					mockRepo.On("UnlinkChatMessages", mock.Anything, int64(1), int64(10)).Return(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := service.UnlinkChatMessages(ctx, 1, 10)

					Expect(err).ToNot(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})
		})

		g.Describe("PlayerIsBanned()", func() {
			g.Describe("Player is banned", func() {
				g.BeforeEach(func() {
					mockRepo.On("PlayerIsBanned", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
						Return(true, 18272, nil)
				})

				g.It("Should not return an error", func() {
					_, _, err := service.PlayerIsBanned(ctx, "", "")

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should return true", func() {
					isBanned, _, err := service.PlayerIsBanned(ctx, "", "")

					Expect(err).To(BeNil())
					Expect(isBanned).To(BeTrue())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should return the correct time remaining", func() {
					_, timeRemaining, err := service.PlayerIsBanned(ctx, "", "")

					Expect(err).To(BeNil())
					Expect(timeRemaining).To(Equal(int64(18272)))
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Player is not banned", func() {
				g.BeforeEach(func() {
					mockRepo.On("PlayerIsBanned", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
						Return(false, 0, nil)
				})

				g.It("Should not return an error", func() {
					_, _, err := service.PlayerIsBanned(ctx, "", "")

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should return false", func() {
					isBanned, _, err := service.PlayerIsBanned(ctx, "", "")

					Expect(err).To(BeNil())
					Expect(isBanned).To(BeFalse())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should return 0 for the time remaining", func() {
					_, timeRemaining, err := service.PlayerIsBanned(ctx, "", "")

					Expect(err).To(BeNil())
					Expect(timeRemaining).To(Equal(int64(0)))
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Repo error", func() {
				g.BeforeEach(func() {
					mockRepo.On("PlayerIsBanned", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
						Return(false, 0, fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, _, err := service.PlayerIsBanned(ctx, "", "")

					Expect(err).ToNot(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})
		})
	})
}
