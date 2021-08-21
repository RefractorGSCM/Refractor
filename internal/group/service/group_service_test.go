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
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"math"
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	var mockRepo *mocks.GroupRepo
	var authorizer *mocks.Authorizer
	var service *groupService
	var ctx = context.TODO()

	g.Describe("User Service", func() {
		g.BeforeEach(func() {
			mockRepo = new(mocks.GroupRepo)
			authorizer = new(mocks.Authorizer)
			service = &groupService{mockRepo, authorizer, time.Second * 2, zap.NewNop()}
		})

		g.Describe("Store()", func() {
			g.Describe("Group stored successfully", func() {
				g.It("Should not return an error", func() {
					mockRepo.On("Store", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)

					err := service.Store(ctx, &domain.Group{Name: "Test Group"})

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.Describe("A group is being stored with the super admin flag set in the permissions field", func() {
					g.BeforeEach(func() {
						mockRepo.On("Store", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)
					})

					g.It("Should unset the super admin flag in the permissions field", func() {
						newGroup := &domain.Group{
							Permissions: bitperms.NewPermissionBuilder().
								AddFlag(perms.GetFlag(perms.FlagSuperAdmin)).
								AddFlag(bitperms.GetFlag(1)).
								AddFlag(bitperms.GetFlag(2)).
								AddFlag(bitperms.GetFlag(3)).
								GetPermission().String(),
						}

						_ = service.Store(context.TODO(), newGroup)

						newVal := newGroup.Permissions
						expected := bitperms.NewPermissionBuilder().
							AddFlag(bitperms.GetFlag(1)).
							AddFlag(bitperms.GetFlag(2)).
							AddFlag(bitperms.GetFlag(3)).
							GetPermission()

						Expect(newVal).To(Equal(expected.String()))
					})
				})
			})
		})

		g.Describe("GetByID()", func() {
			g.Describe("Result fetched successfully", func() {
				g.It("Should not return an error", func() {
					mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("int64")).Return(&domain.Group{}, nil)

					_, err := service.GetByID(ctx, 1)

					Expect(err).To(BeNil())
				})

				g.It("Should return the correct group", func() {
					mockGroup := &domain.Group{
						ID:          1,
						Name:        "Test Group",
						Color:       5423552,
						Position:    15,
						Permissions: "345276874377",
						CreatedAt:   time.Time{},
						ModifiedAt:  time.Time{},
					}

					mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("int64")).Return(mockGroup, nil)

					foundGroup, err := service.GetByID(ctx, 1)

					Expect(err).To(BeNil())
					Expect(foundGroup).To(Equal(mockGroup))
				})
			})
		})

		g.Describe("GetAll()", func() {
			var mockGroups []*domain.Group

			g.BeforeEach(func() {
				mockGroups = []*domain.Group{}

				mockGroups = append(mockGroups, &domain.Group{
					ID:          1,
					Name:        "Test Group",
					Color:       5423552,
					Position:    15,
					Permissions: "345276874377",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				})

				mockGroups = append(mockGroups, &domain.Group{
					ID:          2,
					Name:        "Test Group 2",
					Color:       542355452,
					Position:    14,
					Permissions: "34527324326874377",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				})

				mockGroups = append(mockGroups, &domain.Group{
					ID:          3,
					Name:        "Test Group 3",
					Color:       452355452,
					Position:    6,
					Permissions: "44554645664534434",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				})
			})

			g.Describe("Results fetched successfully", func() {
				var baseGroup *domain.Group

				g.BeforeEach(func() {
					baseGroup = &domain.Group{
						ID:          -1,
						Name:        "Everyone",
						Color:       0xcecece,
						Position:    math.MaxInt32,
						Permissions: "2738628437",
					}

					mockRepo.On("GetBaseGroup", mock.Anything).Return(baseGroup, nil)
				})

				g.It("Should not return an error", func() {
					mockRepo.On("GetAll", mock.Anything).Return([]*domain.Group{}, nil)

					_, err := service.GetAll(ctx)

					Expect(err).To(BeNil())
				})

				g.It("Should return the correct groups", func() {
					mockRepo.On("GetAll", mock.Anything).Return(mockGroups, nil)

					foundGroups, err := service.GetAll(ctx)

					expected := mockGroups
					expected = append(expected, baseGroup)

					Expect(err).To(BeNil())
					Expect(foundGroups).To(Equal(expected))
				})

				g.It("Should return the groups sorted ascendingly by position", func() {
					mockRepo.On("GetAll", mock.Anything).Return(mockGroups, nil)

					expected := mockGroups
					expected = append(mockGroups, baseGroup)
					expected = domain.GroupSlice(expected).SortByPosition()

					foundGroups, err := service.GetAll(ctx)

					Expect(err).To(BeNil())
					Expect(foundGroups).To(Equal(expected))
				})
			})
		})

		g.Describe("Delete()", func() {
			g.Describe("Target group found", func() {
				g.BeforeEach(func() {
					mockRepo.On("Delete", mock.Anything, mock.AnythingOfType("int64")).Return(nil)
				})

				g.It("Should not return an error", func() {
					err := service.Delete(context.TODO(), 1)

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Target group was not found", func() {
				g.It("Should return the domain.ErrNotFound error", func() {
					mockRepo.On("Delete", mock.Anything, mock.AnythingOfType("int64")).Return(domain.ErrNotFound)

					err := service.Delete(context.TODO(), 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					mockRepo.AssertExpectations(t)
				})
			})
		})

		g.Describe("Update()", func() {
			var updatedGroup *domain.Group

			g.BeforeEach(func() {
				updatedGroup = &domain.Group{
					ID:          1,
					Name:        "Updated Group",
					Color:       0xcecece,
					Position:    6,
					Permissions: "7456223",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				}
			})

			g.Describe("Target group found", func() {
				g.BeforeEach(func() {
					mockRepo.On("Update", mock.Anything, mock.AnythingOfType("int64"),
						mock.AnythingOfType("domain.UpdateArgs")).Return(updatedGroup, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.Update(context.TODO(), 1, domain.UpdateArgs{})

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})

				g.It("Should return the updated group", func() {
					updated, err := service.Update(context.TODO(), 1, domain.UpdateArgs{})

					Expect(err).To(BeNil())
					Expect(updated).To(Equal(updatedGroup))
					mockRepo.AssertExpectations(t)
				})

				g.Describe("A new permissions value was provided with the super admin flag set", func() {
					g.It("Should update the args to contain a new permissions string without the super admin flag set", func() {
						args := domain.UpdateArgs{
							"Permissions": bitperms.NewPermissionBuilder().
								AddFlag(perms.GetFlag(perms.FlagSuperAdmin)).
								AddFlag(bitperms.GetFlag(1)).
								AddFlag(bitperms.GetFlag(2)).
								AddFlag(bitperms.GetFlag(3)).
								GetPermission().String(),
						}
						_, _ = service.Update(context.TODO(), 1, args)

						newVal := args["Permissions"].(string)
						expected := bitperms.NewPermissionBuilder().
							AddFlag(bitperms.GetFlag(1)).
							AddFlag(bitperms.GetFlag(2)).
							AddFlag(bitperms.GetFlag(3)).
							GetPermission()

						Expect(newVal).To(Equal(expected.String()))
					})
				})
			})

			g.Describe("Target group was not found", func() {
				g.BeforeEach(func() {
					mockRepo.On("Update", mock.Anything, mock.AnythingOfType("int64"),
						mock.AnythingOfType("domain.UpdateArgs")).Return(nil, domain.ErrNotFound)
				})

				g.It("Should return the domain.ErrNotFound error", func() {
					_, err := service.Update(context.TODO(), 1, domain.UpdateArgs{})

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					mockRepo.AssertExpectations(t)
				})

				g.It("Should return nil as the group", func() {
					g, _ := service.Update(context.TODO(), 1, domain.UpdateArgs{})

					Expect(g).To(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})
		})

		g.Describe("Reorder()", func() {
			g.Describe("Reorder success", func() {
				g.BeforeEach(func() {
					mockRepo.On("Reorder", mock.Anything, mock.AnythingOfType("[]*domain.GroupReorderInfo")).
						Return(nil)
				})

				g.It("Should not return an error", func() {
					err := service.Reorder(context.TODO(), []*domain.GroupReorderInfo{})

					Expect(err).To(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})

			g.Describe("Reorder error", func() {
				g.BeforeEach(func() {
					mockRepo.On("Reorder", mock.Anything, mock.AnythingOfType("[]*domain.GroupReorderInfo")).
						Return(fmt.Errorf(""))
				})

				g.It("Should return an error", func() {
					err := service.Reorder(context.TODO(), []*domain.GroupReorderInfo{})

					Expect(err).ToNot(BeNil())
					mockRepo.AssertExpectations(t)
				})
			})
		})

		g.Describe("UpdateBase()", func() {
			var baseGroup *domain.Group
			var updateArgs domain.UpdateArgs

			g.BeforeEach(func() {
				baseGroup = &domain.Group{
					ID:          -1,
					Name:        "Everyone",
					Color:       0xcecece,
					Position:    math.MaxInt32,
					Permissions: "1",
					CreatedAt:   time.Time{},
					ModifiedAt:  time.Time{},
				}

				color := 0xececec
				permissions := "2"

				updateArgs = domain.UpdateArgs{
					"Color":       &color,
					"Permissions": &permissions,
				}
			})

			g.Describe("Successful update", func() {
				var expected *domain.Group

				g.BeforeEach(func() {
					mockRepo.On("GetBaseGroup", mock.Anything).Return(baseGroup, nil)
					mockRepo.On("SetBaseGroup", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)

					expected = &domain.Group{
						ID:          -1,
						Name:        "Everyone",
						Color:       *updateArgs["Color"].(*int),
						Position:    math.MaxInt32,
						Permissions: *updateArgs["Permissions"].(*string),
						CreatedAt:   time.Time{},
						ModifiedAt:  time.Time{},
					}
				})

				g.It("Should not return an error", func() {
					_, err := service.UpdateBase(context.TODO(), updateArgs)

					Expect(err).To(BeNil())
					mock.AssertExpectationsForObjects(t)
				})

				g.It("Should return the updated group", func() {
					updateArgs["Name"] = "Should not update"
					updateArgs["Position"] = 10

					updated, _ := service.UpdateBase(context.TODO(), updateArgs)

					Expect(updated).To(Equal(expected))
					mock.AssertExpectationsForObjects(t)
				})

				g.It("Should not update the name or position fields", func() {
					updated, _ := service.UpdateBase(context.TODO(), updateArgs)

					Expect(updated).To(Equal(expected))
					mock.AssertExpectationsForObjects(t)
				})
			})

			g.Describe("Repository error", func() {
				g.It("Should return an error on GetBaseGroup repo error", func() {
					mockRepo.On("GetBaseGroup", mock.Anything).Return(nil, fmt.Errorf("getbasegroup error"))

					_, err := service.UpdateBase(context.TODO(), updateArgs)

					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("getbasegroup error"))
				})

				g.It("Should return an error on SetBaseGroup repo error", func() {
					mockRepo.On("GetBaseGroup", mock.Anything).Return(baseGroup, nil)
					mockRepo.On("SetBaseGroup", mock.Anything, mock.AnythingOfType("*domain.Group")).
						Return(fmt.Errorf("setbasegroup error"))

					_, err := service.UpdateBase(context.TODO(), updateArgs)

					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("setbasegroup error"))
				})
			})
		})

		g.Describe("canSetGroup()", func() {
			var groupctx domain.GroupSetContext

			g.BeforeEach(func() {
				groupctx = domain.GroupSetContext{
					SetterUserID: "userid1",
					TargetUserID: "userid2",
					GroupID:      1,
				}
			})

			g.Describe("The setting user is a super admin", func() {
				g.BeforeEach(func() {
					superAdminPerms, _ := bitperms.FromString("1")

					authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(superAdminPerms, nil)
				})

				g.It("Should return true", func() {
					hasPermission, err := service.canSetGroup(ctx, groupctx)

					Expect(err).To(BeNil())
					Expect(hasPermission).To(BeTrue())
				})
			})

			g.Describe("The group being given does not have administrator access", func() {
				g.BeforeEach(func() {
					mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domain.Group{
						ID:          1,
						Name:        "Target Group",
						Color:       0,
						Position:    5,
						Permissions: bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagViewServers)).GetPermission().String(),
					}, nil)
				})

				g.Describe("The setting user is an administrator and the target user is not an admin", func() {
					g.BeforeEach(func() {
						setterPerms := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagAdministrator)).GetPermission()
						authorizer.On("GetPermissions", mock.Anything, mock.Anything, "userid1").Return(setterPerms, nil)

						targetPerms := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagViewServers)).GetPermission()
						authorizer.On("GetPermissions", mock.Anything, mock.Anything, "userid2").Return(targetPerms, nil)
					})

					g.It("Should return true", func() {
						hasPermission, err := service.canSetGroup(ctx, groupctx)

						Expect(err).To(BeNil())
						Expect(hasPermission).To(BeTrue())
					})
				})

				g.Describe("The setting user is not an administrator", func() {
					g.BeforeEach(func() {
						setterPerms := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagViewServers)).GetPermission()
						authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(setterPerms, nil)
					})

					g.It("Should return false", func() {
						hasPermission, err := service.canSetGroup(ctx, groupctx)

						Expect(err).To(BeNil())
						Expect(hasPermission).To(BeFalse())
					})
				})

				g.Describe("Both the setting and target users are administrators", func() {
					g.BeforeEach(func() {
						setterPerms := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagAdministrator)).GetPermission()
						authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(setterPerms, nil).Once()

						targetPerms := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagAdministrator)).GetPermission()
						authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(targetPerms, nil).Once()
					})

					g.It("Should return false", func() {
						hasPermission, err := service.canSetGroup(ctx, groupctx)

						Expect(err).To(BeNil())
						Expect(hasPermission).To(BeFalse())
					})
				})

				g.Describe("The setting user is an administrator and the target is a super admin", func() {
					g.BeforeEach(func() {
						setterPerms := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagAdministrator)).GetPermission()
						authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(setterPerms, nil).Once()

						targetPerms := bitperms.NewPermissionBuilder().AddFlag(perms.GetFlag(perms.FlagSuperAdmin)).GetPermission()
						authorizer.On("GetPermissions", mock.Anything, mock.Anything, mock.Anything).Return(targetPerms, nil).Once()
					})

					g.It("Should return false", func() {
						hasPermission, err := service.canSetGroup(ctx, groupctx)

						Expect(err).To(BeNil())
						Expect(hasPermission).To(BeFalse())
					})
				})
			})
		})

		g.Describe("SetServerOverrides()", func() {
			g.Describe("Success", func() {
				g.BeforeEach(func() {
					mockRepo.On("SetServerOverrides", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.SetServerOverrides(ctx, 1, 1, &domain.Overrides{
						AllowOverrides: "0",
						DenyOverrides:  "0",
					})

					Expect(err).To(BeNil())
				})

				g.Describe("Non server scoped permissions provided", func() {
					g.It("Should not return an error", func() {
						_, err := service.SetServerOverrides(ctx, 1, 1, &domain.Overrides{
							AllowOverrides: "1",
							DenyOverrides:  "2",
						})

						Expect(err).To(BeNil())
					})

					g.It("Should return overrides filtered by server scope", func() {
						ovr, _ := service.SetServerOverrides(ctx, 1, 1, &domain.Overrides{
							AllowOverrides: "1",
							DenyOverrides:  "2",
						})

						Expect(ovr.AllowOverrides).To(Equal("0"))
						Expect(ovr.DenyOverrides).To(Equal("0"))
					})
				})
			})

			g.Describe("Repo error", func() {
				g.BeforeEach(func() {
					mockRepo.On("SetServerOverrides", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := service.SetServerOverrides(ctx, 1, 1, &domain.Overrides{
						AllowOverrides: "0",
						DenyOverrides:  "0",
					})

					Expect(err).ToNot(BeNil())
				})
			})
		})
	})
}
