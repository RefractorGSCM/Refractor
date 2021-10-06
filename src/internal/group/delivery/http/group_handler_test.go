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

package http

import (
	"Refractor/domain"
	"Refractor/domain/mocks"
	"Refractor/params"
	"Refractor/pkg/api"
	"Refractor/pkg/perms"
	"bytes"
	"encoding/json"
	"github.com/franela/goblin"
	"github.com/gorilla/schema"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/gomega"
	kratos "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	g := goblin.Goblin(t)

	// Special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	var logger = zap.NewNop()
	var formEncoder = schema.NewEncoder()
	var e = echo.New()
	e.HTTPErrorHandler = api.GetEchoErrorHandler(logger)

	var service *mocks.GroupService
	var authorizer *mocks.Authorizer
	var handler groupHandler
	var formData url.Values
	var user *domain.AuthUser = &domain.AuthUser{
		Traits: &domain.Traits{
			Email:    "test@test.com",
			Username: "testuser",
		},
		Session: &kratos.Session{
			Identity: kratos.Identity{
				Id: "user-uuid-thingy",
			},
		},
	}

	if formEncoder == nil {

	}

	g.Describe("GetPermissions()", func() {
		g.BeforeEach(func() {
			e.HTTPErrorHandler = api.GetEchoErrorHandler(logger)
			service = new(mocks.GroupService)
			authorizer = new(mocks.Authorizer)
			handler = groupHandler{
				service:    service,
				authorizer: authorizer,
				logger:     logger,
			}
			formData = url.Values{}
		})

		g.Describe("Success", func() {
			var rec *httptest.ResponseRecorder
			var response *domain.Response

			g.BeforeEach(func() {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/permissions", strings.NewReader(formData.Encode()))
				rec = httptest.NewRecorder()
				c := e.NewContext(req, rec)
				c.Set("user", user)

				err := handler.GetPermissions(c)
				Expect(err).To(BeNil())

				err = json.Unmarshal(rec.Body.Bytes(), &response)
				Expect(err).To(BeNil())
			})

			g.It("Should return all permissions", func() {
				var gotPerms []resPermission
				data, err := json.Marshal(response.Payload)
				Expect(err).To(BeNil())
				err = json.Unmarshal(data, &gotPerms)
				Expect(err).To(BeNil())

				Expect(len(gotPerms)).To(Equal(len(perms.GetAll())))
			})

			g.It("Should return with the status code http.StatusOK", func() {
				Expect(rec.Code).To(Equal(http.StatusOK))
			})

			g.It("Should return a response with the success field set to true", func() {
				Expect(response.Success).To(BeTrue())
			})
		})
	})

	g.Describe("CreateGroup()", func() {
		g.BeforeEach(func() {
			e.HTTPErrorHandler = api.GetEchoErrorHandler(logger)
			service = new(mocks.GroupService)
			authorizer = new(mocks.Authorizer)
			handler = groupHandler{
				service:    service,
				authorizer: authorizer,
				logger:     logger,
			}
			formData = url.Values{}
		})

		g.Describe("Success", func() {
			var rec *httptest.ResponseRecorder
			var response *domain.Response
			var body *params.CreateGroupParams

			g.BeforeEach(func() {
				service.On("Store", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)

				body = &params.CreateGroupParams{
					Name:        "Test Group",
					Color:       0xCECECE,
					Position:    1,
					Permissions: "1",
				}

				data, err := json.Marshal(body)
				Expect(err).To(BeNil())

				req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/", bytes.NewReader(data))
				req.Header.Set("Content-Type", "application/json")
				rec = httptest.NewRecorder()
				c := e.NewContext(req, rec)
				c.Set("user", user)

				err = handler.CreateGroup(c)
				Expect(err).To(BeNil())

				err = json.Unmarshal(rec.Body.Bytes(), &response)
				Expect(err).To(BeNil())
			})

			g.It("Should respond with a newly created group", func() {
				var gotGroup *domain.Group
				data, err := json.Marshal(response.Payload)
				Expect(err).To(BeNil())
				err = json.Unmarshal(data, &gotGroup)
				Expect(err).To(BeNil())

				Expect(gotGroup.Name).To(Equal(body.Name))
				Expect(gotGroup.Color).To(Equal(body.Color))
				Expect(gotGroup.Position).To(Equal(body.Position))
				Expect(gotGroup.Permissions).To(Equal(body.Permissions))
			})

			g.It("Should respond with http.StatusCreated", func() {
				Expect(rec.Code).To(Equal(http.StatusCreated))
			})

			g.It("Should return a response with the success field set to true", func() {
				Expect(response.Success).To(BeTrue())
			})
		})

		g.Describe("Input error", func() {
			var rec *httptest.ResponseRecorder
			var body *params.CreateGroupParams
			var httpErr *domain.HTTPError

			var run = func() {
				service.On("Store", mock.Anything, mock.Anything).Return(nil)

				data, err := json.Marshal(body)
				Expect(err).To(BeNil())
				req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/", bytes.NewReader(data))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec = httptest.NewRecorder()
				c := e.NewContext(req, rec)
				c.Set("user", user)

				err = handler.CreateGroup(c)
				Expect(err).ToNot(BeNil())

				var ok bool
				httpErr, ok = err.(*domain.HTTPError)
				Expect(ok).To(BeTrue())
			}

			var resetBody = func() {
				body = &params.CreateGroupParams{
					Name:        "Name",
					Color:       0xCECECE,
					Position:    1,
					Permissions: "1",
				}
			}

			g.Before(func() {
				resetBody()
			})

			g.It("Error should have status set to http.StatusBadRequest", func() {
				body.Name = ""
				run()
				Expect(httpErr.Status).To(Equal(http.StatusBadRequest))
				resetBody()

				body.Name = strings.Repeat("a", 30)
				run()
				Expect(httpErr.Status).To(Equal(http.StatusBadRequest))
				resetBody()

				body.Permissions = ""
				run()
				Expect(httpErr.Status).To(Equal(http.StatusBadRequest))
				resetBody()

				body.Permissions = strings.Repeat("a", 30)
				run()
				Expect(httpErr.Status).To(Equal(http.StatusBadRequest))
				resetBody()

				body.Position = 0
				run()
				Expect(httpErr.Status).To(Equal(http.StatusBadRequest))
				resetBody()

				body.Position = -1
				run()
				Expect(httpErr.Status).To(Equal(http.StatusBadRequest))
				resetBody()

				body.Color = -1
				run()
				Expect(httpErr.Status).To(Equal(http.StatusBadRequest))
				resetBody()

				body.Color = math.MaxInt32
				run()
				Expect(httpErr.Status).To(Equal(http.StatusBadRequest))
				resetBody()
			})
		})
	})
}
