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

	g.Describe("Flagged Word Service", func() {
		var service *flaggedWordService
		var repo *mocks.FlaggedWordRepo
		var logger = zap.NewNop()
		var ctx context.Context

		g.BeforeEach(func() {
			repo = new(mocks.FlaggedWordRepo)
			service = &flaggedWordService{
				repo:    repo,
				timeout: time.Second * 2,
				logger:  logger,
			}
			ctx = context.TODO()
		})

		g.Describe("Store()", func() {
			var flaggedWord *domain.FlaggedWord

			g.BeforeEach(func() {
				flaggedWord = &domain.FlaggedWord{
					ID:   1,
					Word: "flagged word",
				}
			})

			g.Describe("Successful store", func() {
				g.BeforeEach(func() {
					repo.On("Store", mock.Anything, mock.AnythingOfType("*domain.FlaggedWord")).
						Return(nil)
				})

				g.It("Should not return an error", func() {
					err := service.Store(ctx, flaggedWord)

					Expect(err).To(BeNil())
					repo.AssertExpectations(t)
				})
			})

			g.Describe("Repo error", func() {
				g.BeforeEach(func() {
					repo.On("Store", mock.Anything, mock.Anything).Return(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := service.Store(ctx, flaggedWord)

					Expect(err).ToNot(BeNil())
					repo.AssertExpectations(t)
				})
			})
		})

		g.Describe("GetAll()", func() {
			var mockFlaggedWords []*domain.FlaggedWord

			g.BeforeEach(func() {
				mockFlaggedWords = []*domain.FlaggedWord{
					{
						ID:   1,
						Word: "word 1",
					},
					{
						ID:   2,
						Word: "word 2",
					},
					{
						ID:   3,
						Word: "word 3",
					},
					{
						ID:   4,
						Word: "word 4",
					},
				}
			})

			g.Describe("Results were found", func() {
				g.BeforeEach(func() {
					repo.On("GetAll", mock.Anything).Return(mockFlaggedWords, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.GetAll(ctx)

					Expect(err).To(BeNil())
					repo.AssertExpectations(t)
				})

				g.It("Should return the correct results", func() {
					got, err := service.GetAll(ctx)

					Expect(err).To(BeNil())
					Expect(got).To(Equal(mockFlaggedWords))
					repo.AssertExpectations(t)
				})
			})

			g.Describe("No results found", func() {
				g.BeforeEach(func() {
					repo.On("GetAll", mock.Anything).Return(nil, domain.ErrNotFound)
				})

				g.It("Should return a domain.ErrNotFound error", func() {
					_, err := service.GetAll(ctx)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					repo.AssertExpectations(t)
				})

				g.It("Should return nil", func() {
					got, err := service.GetAll(ctx)

					Expect(err).ToNot(BeNil())
					Expect(got).To(BeNil())
					repo.AssertExpectations(t)
				})
			})
		})

		g.Describe("Update()", func() {
			var updated *domain.FlaggedWord

			g.BeforeEach(func() {
				updated = &domain.FlaggedWord{
					ID:   1,
					Word: "updated word",
				}
			})

			g.Describe("Successful update", func() {
				g.BeforeEach(func() {
					repo.On("Update", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("string")).
						Return(updated, nil)
				})

				g.It("Should not return an error", func() {
					_, err := service.Update(ctx, updated.ID, "updated word")

					Expect(err).To(BeNil())
					repo.AssertExpectations(t)
				})

				g.It("Should not return the updated struct", func() {
					got, err := service.Update(ctx, updated.ID, "updated word")

					Expect(err).To(BeNil())
					Expect(got).To(Equal(updated))
					repo.AssertExpectations(t)
				})
			})

			g.Describe("Target not found", func() {
				g.BeforeEach(func() {
					repo.On("Update", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("string")).
						Return(nil, domain.ErrNotFound)
				})

				g.It("Should return a domain.ErrNotfound error", func() {
					_, err := service.Update(ctx, updated.ID, "updated word")

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					repo.AssertExpectations(t)
				})
			})

			g.Describe("Repo error", func() {
				g.BeforeEach(func() {
					repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					_, err := service.Update(ctx, 1, "new word")

					Expect(err).ToNot(BeNil())
					repo.AssertExpectations(t)
				})

				g.It("Should return nil", func() {
					got, err := service.Update(ctx, 1, "new word")

					Expect(err).ToNot(BeNil())
					Expect(got).To(BeNil())
					repo.AssertExpectations(t)
				})
			})
		})

		g.Describe("Delete()", func() {
			g.Describe("Successful delete", func() {
				g.BeforeEach(func() {
					repo.On("Delete", mock.Anything, mock.Anything).Return(nil)
				})

				g.It("Should not return an error", func() {
					err := service.Delete(ctx, 1)

					Expect(err).To(BeNil())
					repo.AssertExpectations(t)
				})
			})

			g.Describe("Target not found", func() {
				g.BeforeEach(func() {
					repo.On("Delete", mock.Anything, mock.Anything).Return(domain.ErrNotFound)
				})

				g.It("Should not return a domain.ErrNotFound error", func() {
					err := service.Delete(ctx, 1)

					Expect(errors.Cause(err)).To(Equal(domain.ErrNotFound))
					repo.AssertExpectations(t)
				})
			})

			g.Describe("Repo error", func() {
				g.BeforeEach(func() {
					repo.On("Delete", mock.Anything, mock.Anything).Return(fmt.Errorf("err"))
				})

				g.It("Should return an error", func() {
					err := service.Delete(ctx, 1)

					Expect(err).ToNot(BeNil())
					repo.AssertExpectations(t)
				})
			})
		})
	})
}
