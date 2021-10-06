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

package authorizer

import (
	"Refractor/domain"
	"Refractor/domain/mocks"
	"Refractor/pkg/bitperms"
	"context"
	"fmt"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"math"
	"testing"
	"time"
)

func TestAuthorizer(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	var testAuthChecker = func(permissions *bitperms.Permissions) (bool, error) {
		return true, nil
	}

	g.Describe("Test Wrapper", func() {
		var _baseGroup *domain.Group
		var _userGroups []*domain.Group
		var _userOverrides *domain.Overrides

		g.BeforeEach(func() {

			// base group (everyone) setup
			baseGroupPerms := bitperms.NewPermissionBuilder().
				AddFlag(bitperms.GetFlag(0)).
				GetPermission()

			_baseGroup = &domain.Group{
				ID:          1, // BASE GROUP MUST BE ID 1
				Name:        "Everyone",
				Color:       1234,
				Position:    math.MaxInt32,
				Permissions: baseGroupPerms.String(),
			}

			// extra groups setup
			groupPerms := bitperms.NewPermissionBuilder().
				AddFlag(bitperms.GetFlag(0)).
				AddFlag(bitperms.GetFlag(1)).
				GetPermission()

			_userGroups = []*domain.Group{}
			_userGroups = append(_userGroups, &domain.Group{
				ID:          2,
				Name:        "Group 2",
				Color:       1234,
				Position:    4,
				Permissions: groupPerms.String(),
			})

			groupPerms = bitperms.NewPermissionBuilder().
				AddFlag(bitperms.GetFlag(1)).
				AddFlag(bitperms.GetFlag(3)).
				AddFlag(bitperms.GetFlag(4)).
				GetPermission()

			_userGroups = append(_userGroups, &domain.Group{
				ID:          3,
				Name:        "Group 3",
				Color:       1234,
				Position:    3,
				Permissions: groupPerms.String(),
			})

			// user overrides setup
			denyOverPerms := bitperms.NewPermissionBuilder().
				AddFlag(bitperms.GetFlag(1)).
				AddFlag(bitperms.GetFlag(2)).
				AddFlag(bitperms.GetFlag(3)).
				GetPermission()

			allowOverPerms := bitperms.NewPermissionBuilder().
				AddFlag(bitperms.GetFlag(2)).
				AddFlag(bitperms.GetFlag(5)).
				GetPermission()

			_userOverrides = &domain.Overrides{
				AllowOverrides: allowOverPerms.String(),
				DenyOverrides:  denyOverPerms.String(),
			}
		})

		g.Describe("HasPermission()", func() {
			var repo *mocks.GroupRepo
			var a *authorizer

			g.BeforeEach(func() {
				repo = new(mocks.GroupRepo)

				a = &authorizer{
					groupRepo: repo,
				}

				repo.On("GetBaseGroup", mock.Anything).Return(_baseGroup, nil)
				repo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(_userGroups, nil)
				repo.On("GetUserOverrides", mock.Anything, mock.AnythingOfType("string")).Return(_userOverrides, nil)
				repo.On("GetServerOverrides", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			})

			g.Describe("A valid AuthScope was provided", func() {
				g.It("Should not return an error", func() {
					var as domain.AuthScope
					var err error

					// Refractor auth scope
					as = domain.AuthScope{
						Type: domain.AuthObjRefractor,
						ID:   nil,
					}
					_, err = a.HasPermission(context.TODO(), as, "userID", testAuthChecker)
					Expect(err).To(BeNil())

					// Server auth scope
					as = domain.AuthScope{
						Type: domain.AuthObjServer,
						ID:   int64(1),
					}
					_, err = a.HasPermission(context.TODO(), as, "userID", testAuthChecker)
					Expect(err).To(BeNil())
				})
			})

			g.Describe("An invalid AuthScope was provided", func() {
				g.It("Should return an error", func() {
					as := domain.AuthScope{
						Type: "Invalid",
					}

					_, err := a.HasPermission(context.TODO(), as, "userID", testAuthChecker)

					Expect(err).ToNot(BeNil())
				})
			})
		})

		g.Describe("computePermissionsRefractor()", func() {
			var repo *mocks.GroupRepo
			var a *authorizer

			g.BeforeEach(func() {
				repo = &mocks.GroupRepo{}

				a = &authorizer{
					groupRepo: repo,
				}
			})

			g.Describe("Permissions computed successfully", func() {
				var baseGroup *domain.Group
				var userGroups []*domain.Group

				g.BeforeEach(func() {
					// base group (everyone) setup
					baseGroupPerms := bitperms.NewPermissionBuilder().
						AddFlag(bitperms.GetFlag(0)).
						GetPermission()

					baseGroup = &domain.Group{
						ID:          1, // BASE GROUP MUST BE ID 1
						Name:        "Everyone",
						Color:       1234,
						Position:    math.MaxInt32,
						Permissions: baseGroupPerms.String(),
					}

					repo.On("GetBaseGroup", mock.Anything).Return(baseGroup, nil)

					// extra groups setup
					groupPerms := bitperms.NewPermissionBuilder().
						AddFlag(bitperms.GetFlag(0)).
						AddFlag(bitperms.GetFlag(1)).
						GetPermission()

					userGroups = []*domain.Group{}
					userGroups = append(userGroups, &domain.Group{
						ID:          2,
						Name:        "Group 2",
						Color:       1234,
						Position:    4,
						Permissions: groupPerms.String(),
					})

					groupPerms = bitperms.NewPermissionBuilder().
						AddFlag(bitperms.GetFlag(1)).
						AddFlag(bitperms.GetFlag(3)).
						AddFlag(bitperms.GetFlag(4)).
						GetPermission()

					userGroups = append(userGroups, &domain.Group{
						ID:          3,
						Name:        "Group 3",
						Color:       1234,
						Position:    3,
						Permissions: groupPerms.String(),
					})

					repo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(userGroups, nil)

					// user overrides setup
					denyOverPerms := bitperms.NewPermissionBuilder().
						AddFlag(bitperms.GetFlag(1)).
						AddFlag(bitperms.GetFlag(2)).
						AddFlag(bitperms.GetFlag(3)).
						GetPermission()

					allowOverPerms := bitperms.NewPermissionBuilder().
						AddFlag(bitperms.GetFlag(2)).
						AddFlag(bitperms.GetFlag(5)).
						GetPermission()

					_userOverrides = &domain.Overrides{
						AllowOverrides: allowOverPerms.String(),
						DenyOverrides:  denyOverPerms.String(),
					}

					repo.On("GetUserOverrides", mock.Anything, mock.AnythingOfType("string")).Return(_userOverrides, nil)

					// OVERALL, this is what the above code does:
					// (s0 means 1 << 0, s1 means 1 << 1, etc. determined using bitperms.GetStep)
					// + represents a granted flag, - represents a denied flag
					//
					// Legend->                                         | s0 | s1 | s2 | s3 | s4 | s5 |
					//---------------------------------------------------------------------------------
					// 1.  Apply the following flags to the base group: |+s0 |    |    |    |    |    |
					// 2a. Apply the following flags to group 1: 		|+s0 |+s1 |    |    |    |    |
					// 2b. Apply the following flags to group 2: 		|    |+s1 |    |+s3 |+s4 |    |
					// 3.  Apply the following deny overrides to user:  |    |-s1 |-s2 |-s3 |    |    |
					// 4.  Apply the following allow overrides to user:	|    |    |+s2 |    |    |+s5 |
					//---------------------------------------------------------------------------------
					// Given the above, you can see that the resulting flags are as follows:
					//
					//                                                  | s0 |    | s2 |    | s4 | s5 |
					//
					// Meaning the resulting permissions flag will be:
					//													= (1<<0) | (1<<2) | (1<<4) | (1<<5)
					// 													= 53
				})

				g.It("Should not return an error", func() {
					_, err := a.computePermissionsRefractor(context.TODO(), "userid")

					Expect(err).To(BeNil())
					repo.AssertExpectations(t)
				})

				g.It("Should return the properly computed permissions value", func() {
					computed, _ := a.computePermissionsRefractor(context.TODO(), "userid")

					expected := bitperms.NewPermissionBuilder().
						AddFlag(bitperms.GetFlag(0)).
						AddFlag(bitperms.GetFlag(2)).
						AddFlag(bitperms.GetFlag(4)).
						AddFlag(bitperms.GetFlag(5)).
						GetPermission().Value()

					fmt.Printf("Computed Value: %.12b\n", computed.Value())
					fmt.Printf("Expected Value: %.12b\n", expected)

					Expect(computed.Value()).To(Equal(expected))
				})
			})

			g.Describe("No user groups are set", func() {
				g.BeforeEach(func() {
					repo.On("GetBaseGroup", mock.Anything).Return(_baseGroup, nil)
					repo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(nil, domain.ErrNotFound)
					repo.On("GetUserOverrides", mock.Anything, mock.AnythingOfType("string")).Return(_userOverrides, nil)
				})

				g.It("Should skip checking group permissions", func() {
					bgp, _ := bitperms.FromString(_baseGroup.Permissions)
					bgp, _ = bgp.ComputeDenyOverrides(_userOverrides.DenyOverrides)
					bgp, _ = bgp.ComputeAllowOverrides(_userOverrides.AllowOverrides)

					computed, _ := a.computePermissionsRefractor(context.TODO(), "userid")

					//fmt.Printf("Expected1:    %.10b\n", bgp.Value())
					//fmt.Printf("Computed1:    %.10b\n", computed.Value())

					Expect(computed.Value()).To(Equal(bgp.Value()))
					mock.AssertExpectationsForObjects(t)
				})

				g.It("Should not throw an error", func() {
					_, err := a.computePermissionsRefractor(context.TODO(), "userid")

					Expect(err).To(BeNil())
					mock.AssertExpectationsForObjects(t)
				})
			})

			g.Describe("No user overrides are set", func() {
				g.BeforeEach(func() {
					repo.On("GetBaseGroup", mock.Anything).Return(_baseGroup, nil)
					repo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(_userGroups, nil)
					repo.On("GetUserOverrides", mock.Anything, mock.AnythingOfType("string")).Return(nil, domain.ErrNotFound)
				})

				g.It("Should skip checking override permissions", func() {
					bgp, _ := bitperms.FromString(_baseGroup.Permissions)
					for _, group := range _userGroups {
						perms, _ := bitperms.FromString(group.Permissions)
						bgp = bgp.Or(perms)
					}

					computed, _ := a.computePermissionsRefractor(context.TODO(), "userid")

					//fmt.Printf("Expected1:    %.10b\n", bgp.Value())
					//fmt.Printf("Computed1:    %.10b\n", computed.Value())

					Expect(computed.Value()).To(Equal(bgp.Value()))
					mock.AssertExpectationsForObjects(t)
				})

				g.It("Should not throw an error", func() {
					_, err := a.computePermissionsRefractor(context.TODO(), "userid")

					Expect(err).To(BeNil())
					mock.AssertExpectationsForObjects(t)
				})
			})
		})

		/////////////////////////////////////////////////////////////////////////////////////////////////////
		g.Describe("computePermissionsServer()", func() {
			var repo *mocks.GroupRepo
			var a *authorizer

			g.BeforeEach(func() {
				repo = &mocks.GroupRepo{}

				a = &authorizer{
					groupRepo: repo,
				}
			})

			g.Describe("Permissions computed successfully", func() {
				var baseGroup *domain.Group
				var userGroups []*domain.Group
				var serverID int64

				g.BeforeEach(func() {
					serverID = 1

					// base group (everyone) setup
					baseGroupPerms := bitperms.NewPermissionBuilder().
						AddFlag(bitperms.GetFlag(0)).
						AddFlag(bitperms.GetFlag(7)).
						GetPermission()

					baseGroup = &domain.Group{
						ID:          1, // BASE GROUP MUST BE ID 1
						Name:        "Everyone",
						Color:       1234,
						Position:    math.MaxInt32,
						Permissions: baseGroupPerms.String(),
					}

					repo.On("GetBaseGroup", mock.Anything).Return(baseGroup, nil)
				})

				g.Describe("Everything is present", func() {
					g.BeforeEach(func() {
						// extra groups setup
						groupPerms := bitperms.NewPermissionBuilder().
							AddFlag(bitperms.GetFlag(0)).
							AddFlag(bitperms.GetFlag(1)).
							GetPermission()

						userGroups = []*domain.Group{}
						userGroups = append(userGroups, &domain.Group{
							ID:          2,
							Name:        "Group 2",
							Color:       1234,
							Position:    4,
							Permissions: groupPerms.String(),
						})

						groupPerms = bitperms.NewPermissionBuilder().
							AddFlag(bitperms.GetFlag(1)).
							AddFlag(bitperms.GetFlag(3)).
							AddFlag(bitperms.GetFlag(4)).
							AddFlag(bitperms.GetFlag(5)).
							GetPermission()

						userGroups = append(userGroups, &domain.Group{
							ID:          3,
							Name:        "Group 3",
							Color:       1234,
							Position:    3,
							Permissions: groupPerms.String(),
						})

						repo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(userGroups, nil)

						// group server overrides setup
						groupOverrides := map[int64]*domain.Overrides{
							userGroups[0].ID: &domain.Overrides{
								AllowOverrides: bitperms.NewPermissionBuilder().
									AddFlag(bitperms.GetFlag(1)).
									AddFlag(bitperms.GetFlag(5)).
									AddFlag(bitperms.GetFlag(6)).
									GetPermission().String(),
								DenyOverrides: bitperms.NewPermissionBuilder().
									AddFlag(bitperms.GetFlag(5)).
									AddFlag(bitperms.GetFlag(6)).
									AddFlag(bitperms.GetFlag(0)).
									AddFlag(bitperms.GetFlag(7)).
									GetPermission().String(),
							},
						}

						repo.On("GetServerOverrides", mock.Anything, serverID, userGroups[0].ID).Return(groupOverrides[userGroups[0].ID], nil)
						repo.On("GetServerOverrides", mock.Anything, serverID, userGroups[1].ID).Return(nil, nil)

						// user overrides setup
						denyOverPerms := bitperms.NewPermissionBuilder().
							AddFlag(bitperms.GetFlag(2)).
							AddFlag(bitperms.GetFlag(3)).
							GetPermission()

						allowOverPerms := bitperms.NewPermissionBuilder().
							AddFlag(bitperms.GetFlag(1)).
							AddFlag(bitperms.GetFlag(2)).
							AddFlag(bitperms.GetFlag(5)).
							GetPermission()

						_userOverrides = &domain.Overrides{
							AllowOverrides: allowOverPerms.String(),
							DenyOverrides:  denyOverPerms.String(),
						}

						repo.On("GetUserOverrides", mock.Anything, mock.AnythingOfType("string")).Return(_userOverrides, nil)

						// Here is a visual representation of the changes made to the computed permissions at each step of
						// the server scoped permission computation.
						// 1 means that a flag is on. If it goes from 1 to 0, it means the current step turned it off.
						//                    |            Flags              |
						// Step               | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
						// ----------------------------------------------------
						// Base group perms   | 1 |   |   |   |   |   |   | 1 |
						// Group 2 perms      | 1 | 1 |   |   |   |   |   | 1 |
						// Group 3 perms      | 1 | 1 |   | 1 | 1 | 1 |   | 1 |
						// Group 2 deny ovr.  | 0 | 1 |   | 1 | 1 | 0 |   | 0 |
						// Group 2 allow ovr. |   | 1 |   | 1 | 1 | 1 | 1 |   |
						// User deny ovr.     |   | 1 |   | 0 | 1 | 1 | 1 |   |
						// User allow ovr.    |   | 1 | 1 |   | 1 | 1 | 1 |   |
						// ----------------------------------------------------
						// Final on flags:          1   2       4   5   6
					})

					g.It("Should not return an error", func() {
						_, err := a.computePermissionsServer(context.TODO(), "userid", serverID)

						Expect(err).To(BeNil())
						repo.AssertExpectations(t)
					})

					g.It("Should return the properly computed permissions value", func() {
						computed, _ := a.computePermissionsServer(context.TODO(), "userid", serverID)

						expected := bitperms.NewPermissionBuilder().
							AddFlag(bitperms.GetFlag(1)).
							AddFlag(bitperms.GetFlag(2)).
							AddFlag(bitperms.GetFlag(4)).
							AddFlag(bitperms.GetFlag(5)).
							AddFlag(bitperms.GetFlag(6)).
							GetPermission().Value()

						fmt.Printf("Computed Value: %.12b\n", computed.Value())
						fmt.Printf("Expected Value: %.12b\n", expected)

						Expect(computed.Value()).To(Equal(expected))
					})
				})

				g.Describe("No user groups found", func() {
					g.BeforeEach(func() {
						repo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(nil, domain.ErrNotFound)
						repo.On("GetUserOverrides", mock.Anything, mock.Anything).Return(&domain.Overrides{
							AllowOverrides: "0",
							DenyOverrides:  "0",
						}, nil)
					})

					g.It("Should not return an error", func() {
						_, err := a.computePermissionsServer(context.TODO(), "userid", serverID)

						Expect(err).To(BeNil())
					})

					// We don't test for specific permission outputs since these were already tested for in the
					// "permissions computed successfully" describe block above. We only want to make sure it doesn't
					// error out if something is missing.
				})

				g.Describe("No server overrides", func() {
					g.BeforeEach(func() {
						_userGroups = []*domain.Group{
							{
								ID:          1,
								Name:        "Test",
								Color:       0,
								Position:    1,
								Permissions: "3436278",
								CreatedAt:   time.Time{},
								ModifiedAt:  time.Time{},
							},
						}

						repo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(_userGroups, nil)
						repo.On("GetServerOverrides", mock.Anything, mock.Anything, mock.Anything).Return(nil, domain.ErrNotFound)
						repo.On("GetUserOverrides", mock.Anything, mock.Anything).Return(nil, nil)
					})

					g.It("Should not return an error", func() {
						_, err := a.computePermissionsServer(context.TODO(), "userid", serverID)

						Expect(err).To(BeNil())
					})

					// We don't test for specific permission outputs since these were already tested for in the
					// "permissions computed successfully" describe block above. We only want to make sure it doesn't
					// error out if something is missing.
				})

				g.Describe("No user overrides", func() {
					g.BeforeEach(func() {
						repo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(_userGroups, nil)
						repo.On("GetServerOverrides", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
						repo.On("GetUserOverrides", mock.Anything, mock.Anything).Return(nil, domain.ErrNotFound)
					})

					g.It("Should not return an error", func() {
						_, err := a.computePermissionsServer(context.TODO(), "userid", serverID)

						Expect(err).To(BeNil())
					})

					// We don't test for specific permission outputs since these were already tested for in the
					// "permissions computed successfully" describe block above. We only want to make sure it doesn't
					// error out if something is missing.
				})
			})
		})

		g.Describe("hasPermissionRefractor()", func() {
			var repo *mocks.GroupRepo
			var a *authorizer

			g.BeforeEach(func() {
				repo = new(mocks.GroupRepo)

				a = &authorizer{
					groupRepo: repo,
				}
			})

			g.Describe("NewUser has permission", func() {
				g.BeforeEach(func() {
					repo.On("GetBaseGroup", mock.Anything).Return(_baseGroup, nil)
					repo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(_userGroups, nil)
					repo.On("GetUserOverrides", mock.Anything, mock.AnythingOfType("string")).Return(_userOverrides, nil)
				})

				g.It("Does not return an error", func() {
					_, err := a.hasPermissionRefractor(context.TODO(), "userid", testAuthChecker)

					Expect(err).To(BeNil())
					mock.AssertExpectationsForObjects(t)
				})

				g.It("Returns true", func() {
					hasPermission, _ := a.hasPermissionRefractor(context.TODO(), "userid", testAuthChecker)

					Expect(hasPermission).To(BeTrue())
					mock.AssertExpectationsForObjects(t)
				})
			})

			g.Describe("NewUser does not have permission", func() {
				g.BeforeEach(func() {
					baseGroupPerms := bitperms.NewPermissionBuilder().
						AddFlag(bitperms.GetFlag(60)).
						AddFlag(bitperms.GetFlag(61)).
						AddFlag(bitperms.GetFlag(62)).
						GetPermission()

					_baseGroup.Permissions = baseGroupPerms.String()

					_userGroups = []*domain.Group{
						{
							ID:       2,
							Name:     "Test Group",
							Color:    0,
							Position: 0,
							Permissions: bitperms.NewPermissionBuilder().
								AddFlag(bitperms.GetFlag(60)).
								AddFlag(bitperms.GetFlag(61)).
								AddFlag(bitperms.GetFlag(62)).
								AddFlag(bitperms.GetFlag(63)).
								AddFlag(bitperms.GetFlag(64)).
								AddFlag(bitperms.GetFlag(65)).
								GetPermission().String(),
							CreatedAt:  time.Time{},
							ModifiedAt: time.Time{},
						},
					}

					_userOverrides = &domain.Overrides{
						DenyOverrides: bitperms.NewPermissionBuilder().
							AddFlag(bitperms.GetFlag(61)).
							AddFlag(bitperms.GetFlag(62)).
							AddFlag(bitperms.GetFlag(63)). // target for deny test
							AddFlag(bitperms.GetFlag(64)). // target for deny test
							GetPermission().String(),
						AllowOverrides: bitperms.NewPermissionBuilder().
							AddFlag(bitperms.GetFlag(66)).
							AddFlag(bitperms.GetFlag(62)).
							GetPermission().String(),
					}

					repo.On("GetBaseGroup", mock.Anything).Return(_baseGroup, nil)
					repo.On("GetUserGroups", mock.Anything, mock.AnythingOfType("string")).Return(_userGroups, nil)
					repo.On("GetUserOverrides", mock.Anything, mock.AnythingOfType("string")).Return(_userOverrides, nil)
				})

				g.It("Returns false", func() {
					authChecker := func(permissions *bitperms.Permissions) (bool, error) {
						if !permissions.CheckFlag(bitperms.GetFlag(63)) ||
							!permissions.CheckFlag(bitperms.GetFlag(64)) {
							return false, nil
						}

						return true, nil
					}

					hasPermission, _ := a.hasPermissionRefractor(context.TODO(), "userid", authChecker)

					Expect(hasPermission).To(BeFalse())
					mock.AssertExpectationsForObjects(t)
				})
			})
		})
	})
}
