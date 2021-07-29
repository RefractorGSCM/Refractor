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
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("CreateUser()", func() {
		var repo *mocks.AuthRepo
		var metaRepo *mocks.UserMetaRepo
		var mailService *mocks.MailService
		var service domain.AuthService
		var newUserTraits *domain.Traits
		var newUser *domain.AuthUser

		g.BeforeEach(func() {
			repo = new(mocks.AuthRepo)
			metaRepo = new(mocks.UserMetaRepo)
			mailService = new(mocks.MailService)
			service = NewAuthService(repo, metaRepo, mailService, time.Second*2, zap.NewNop())

			newUserTraits = &domain.Traits{
				Email:    "test@test.com",
				Username: "test",
			}

			newUser = &domain.AuthUser{
				Traits: newUserTraits,
				Session: &kratos.Session{
					Identity: kratos.Identity{
						Id: "testuserid",
					},
				},
			}
		})

		g.Describe("Successful creation", func() {
			g.BeforeEach(func() {
				repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.Traits")).Return(newUser, nil)
				metaRepo.On("Store", mock.Anything, mock.Anything).Return(nil)
				repo.On("GetRecoveryLink", mock.Anything, mock.AnythingOfType("string")).Return("fakelink", nil)
				mailService.On("SendWelcomeEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			})

			g.It("Should not return an error", func() {
				_, err := service.CreateUser(context.TODO(), newUserTraits, "system")

				Expect(err).To(BeNil())
				mailService.AssertExpectations(t)
				repo.AssertExpectations(t)
				metaRepo.AssertExpectations(t)
			})

			g.It("Should return the new user", func() {
				user, _ := service.CreateUser(context.TODO(), newUserTraits, "system")

				Expect(user).To(Equal(newUser))
				mailService.AssertExpectations(t)
				repo.AssertExpectations(t)
				metaRepo.AssertExpectations(t)
			})
		})

		g.Describe("Auth repo CreateUser() error", func() {
			g.BeforeEach(func() {
				repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.Traits")).Return(nil, fmt.Errorf("err"))
			})

			g.It("Should return an error", func() {
				_, err := service.CreateUser(context.TODO(), newUserTraits, "system")

				Expect(err).ToNot(BeNil())
				repo.AssertExpectations(t)
			})
		})

		g.Describe("UserMeta repo error", func() {
			g.BeforeEach(func() {
				repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.Traits")).Return(newUser, nil)
				metaRepo.On("Store", mock.Anything, mock.Anything).Return(fmt.Errorf("err"))
			})

			g.It("Should return an error", func() {
				_, err := service.CreateUser(context.TODO(), newUserTraits, "system")

				Expect(err).ToNot(BeNil())
				repo.AssertExpectations(t)
			})
		})

		g.Describe("Auth repo GetRecoveryLink() error", func() {
			g.BeforeEach(func() {
				repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.Traits")).Return(newUser, nil)
				metaRepo.On("Store", mock.Anything, mock.Anything).Return(nil)
				repo.On("GetRecoveryLink", mock.Anything, mock.AnythingOfType("string")).Return("", fmt.Errorf("err"))
			})

			g.It("Should return an error", func() {
				_, err := service.CreateUser(context.TODO(), newUserTraits, "system")

				Expect(err).ToNot(BeNil())
				repo.AssertExpectations(t)
				metaRepo.AssertExpectations(t)
			})
		})

		g.Describe("Mail service SendWelcomeEmail() error", func() {
			g.BeforeEach(func() {
				repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.Traits")).Return(newUser, nil)
				metaRepo.On("Store", mock.Anything, mock.Anything).Return(nil)
				repo.On("GetRecoveryLink", mock.Anything, mock.AnythingOfType("string")).Return("fakelink", nil)
				mailService.On("SendWelcomeEmail", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("err"))
			})

			g.It("Should return an error", func() {
				_, err := service.CreateUser(context.TODO(), newUserTraits, "system")

				Expect(err).ToNot(BeNil())
				mailService.AssertExpectations(t)
				repo.AssertExpectations(t)
				metaRepo.AssertExpectations(t)
			})
		})
	})
}
