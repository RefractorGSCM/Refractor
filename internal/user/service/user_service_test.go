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
	"context"
	"fmt"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	kratos "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	ctx := context.TODO()

	g.Describe("User Service", func() {
		var metaRepo *mocks.UserMetaRepo
		var groupRepo *mocks.GroupRepo
		var authRepo *mocks.AuthRepo
		var authorizer *mocks.Authorizer
		var service domain.UserService

		g.BeforeEach(func() {
			metaRepo = new(mocks.UserMetaRepo)
			groupRepo = new(mocks.GroupRepo)
			authRepo = new(mocks.AuthRepo)
			authorizer = new(mocks.Authorizer)
			service = NewUserService(metaRepo, authRepo, groupRepo, authorizer, time.Second*2, zap.NewNop())
		})

		g.Describe("GetAllUsers()", func() {
			g.Describe("Users retrieved successfully", func() {
				var authUsers []*domain.AuthUser
				var userGroups []*domain.Group
				var userMeta *domain.UserMeta
				var permVal *bitperms.Permissions

				g.BeforeEach(func() {
					authUsers = []*domain.AuthUser{
						{
							Traits: &domain.Traits{
								Username: "username-1",
								Email:    "username-1@refractor.local",
							},
							Session: &kratos.Session{
								Identity: kratos.Identity{
									Id: "userid-1",
								},
							},
						},
						{
							Traits: &domain.Traits{
								Username: "username-2",
								Email:    "username-2@refractor.local",
							},
							Session: &kratos.Session{
								Identity: kratos.Identity{
									Id: "userid-2",
								},
							},
						},
						{
							Traits: &domain.Traits{
								Username: "username-3",
								Email:    "username-3@refractor.local",
							},
							Session: &kratos.Session{
								Identity: kratos.Identity{
									Id: "userid-3",
								},
							},
						},
						{
							Traits: &domain.Traits{
								Username: "username-4",
								Email:    "username-4@refractor.local",
							},
							Session: &kratos.Session{
								Identity: kratos.Identity{
									Id: "userid-4",
								},
							},
						},
					}

					userGroups = []*domain.Group{
						{
							ID:          1,
							Name:        "Group 1",
							Color:       0xcecece,
							Position:    1,
							Permissions: "1",
						},
						{
							ID:          2,
							Name:        "Group 2",
							Color:       0xececec,
							Position:    2,
							Permissions: "2",
						},
					}

					userMeta = &domain.UserMeta{
						ID:              "userid",
						InitialUsername: "initial-username",
						Username:        "new-username",
						Deactivated:     true,
					}

					permVal, _ = bitperms.FromString("1")

					authRepo.On("GetAllUsers", mock.Anything).Return(authUsers, nil)
					authorizer.On("GetPermissions", mock.Anything, mock.AnythingOfType("domain.AuthScope"),
						mock.AnythingOfType("string")).Return(permVal, nil)
					metaRepo.On("GetByID", mock.Anything, mock.Anything).Return(userMeta, nil)
					groupRepo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(userGroups, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.GetAllUsers(ctx)

					Expect(err).To(BeNil())
					metaRepo.AssertExpectations(t)
					authRepo.AssertExpectations(t)
					groupRepo.AssertExpectations(t)
					authorizer.AssertExpectations(t)
				})

				g.It("Should return the expected list of users", func() {
					var expected []*domain.User

					for _, au := range authUsers {
						usr := &domain.User{
							ID:          au.Identity.Id,
							Username:    au.Traits.Username,
							Permissions: permVal.String(),
							Groups:      userGroups,
							UserMeta:    userMeta,
						}

						expected = append(expected, usr)
					}

					users, err := service.GetAllUsers(ctx)

					Expect(err).To(BeNil())
					Expect(users).To(Equal(expected))
					metaRepo.AssertExpectations(t)
					authRepo.AssertExpectations(t)
					groupRepo.AssertExpectations(t)
					authorizer.AssertExpectations(t)
				})
			})

			g.Describe("Error(s) occurred", func() {
				var authUsers []*domain.AuthUser
				var permVal *bitperms.Permissions
				var groups []*domain.Group

				g.BeforeEach(func() {
					authUsers = []*domain.AuthUser{
						{
							Traits: &domain.Traits{
								Username: "username-1",
								Email:    "username-1@refractor.local",
							},
							Session: &kratos.Session{
								Identity: kratos.Identity{
									Id: "userid-1",
								},
							},
						},
					}

					groups = []*domain.Group{
						{
							ID: 1,
						},
					}

					permVal, _ = bitperms.FromString("1")
				})

				g.Describe("Auth repo error", func() {
					g.BeforeEach(func() {
						authRepo.On("GetAllUsers", mock.Anything).Return(nil, fmt.Errorf("err"))
					})

					g.It("Should return an error", func() {
						_, err := service.GetAllUsers(ctx)

						Expect(err).ToNot(BeNil())
						authRepo.AssertExpectations(t)
					})
				})

				g.Describe("Authorizer error", func() {
					g.BeforeEach(func() {
						authRepo.On("GetAllUsers", mock.Anything).Return(authUsers, nil)
						authorizer.On("GetPermissions", mock.Anything, mock.AnythingOfType("domain.AuthScope"),
							mock.AnythingOfType("string")).Return(nil, fmt.Errorf("err"))
					})

					g.It("Should return an error", func() {
						_, err := service.GetAllUsers(ctx)

						Expect(err).ToNot(BeNil())
						authRepo.AssertExpectations(t)
						authorizer.AssertExpectations(t)
					})
				})

				g.Describe("Group repo error", func() {
					g.BeforeEach(func() {
						authRepo.On("GetAllUsers", mock.Anything).Return(authUsers, nil)
						authorizer.On("GetPermissions", mock.Anything, mock.AnythingOfType("domain.AuthScope"),
							mock.AnythingOfType("string")).Return(permVal, nil)
						groupRepo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(nil, fmt.Errorf("err"))
					})

					g.It("Should return an error", func() {
						_, err := service.GetAllUsers(ctx)

						Expect(err).ToNot(BeNil())
						authRepo.AssertExpectations(t)
						authorizer.AssertExpectations(t)
						groupRepo.AssertExpectations(t)
					})
				})

				g.Describe("Meta repo error", func() {
					g.BeforeEach(func() {
						authRepo.On("GetAllUsers", mock.Anything).Return(authUsers, nil)
						authorizer.On("GetPermissions", mock.Anything, mock.AnythingOfType("domain.AuthScope"),
							mock.AnythingOfType("string")).Return(permVal, nil)
						groupRepo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(groups, nil)
						metaRepo.On("GetByID", mock.Anything, mock.AnythingOfType("string")).Return(nil, fmt.Errorf("err"))
					})

					g.It("Should return an error", func() {
						_, err := service.GetAllUsers(ctx)

						Expect(err).ToNot(BeNil())
						authRepo.AssertExpectations(t)
						authorizer.AssertExpectations(t)
						groupRepo.AssertExpectations(t)
						metaRepo.AssertExpectations(t)
					})
				})
			})
		})
	})
}
