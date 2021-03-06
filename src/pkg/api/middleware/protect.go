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

package middleware

import (
	"Refractor/domain"
	"Refractor/pkg/conf"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	kratos "github.com/ory/kratos-client-go"
	"github.com/pkg/errors"
	"net/http"
)

type res struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var unauthorized = &res{
	Status:  http.StatusUnauthorized,
	Message: "Unauthorized",
}

func NewBrowserProtectMiddleware(config *conf.Config, repo domain.UserMetaRepo) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sessionCookie, err := c.Cookie("ory_kratos_session")
			if err != nil || sessionCookie == nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/k/login")
			}

			toSessionURL := fmt.Sprintf("%s/sessions/whoami", config.KratosPublic)

			req, err := http.NewRequest("GET", toSessionURL, nil)
			if err != nil {
				return err
			}

			for _, cookie := range c.Cookies() {
				req.AddCookie(cookie)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}

			if resp.StatusCode != http.StatusOK {
				return c.Redirect(http.StatusTemporaryRedirect, "/k/login")
			}

			session := &kratos.Session{}
			if err := json.NewDecoder(resp.Body).Decode(session); err != nil {
				return err
			}

			traitBytes, err := json.Marshal(session.Identity.Traits)
			if err != nil {
				return err
			}

			traits := &domain.Traits{}
			if err := json.Unmarshal(traitBytes, traits); err != nil {
				return err
			}

			user := &domain.AuthUser{
				Traits:  traits,
				Session: session,
			}

			ctx := c.Request().Context()

			// Check if the user is deactivated
			isDeactivated, err := repo.IsDeactivated(ctx, user.Identity.Id)
			if err != nil {
				return err
			}

			// If they are deactivated, deny access.
			if isDeactivated {
				return c.HTML(http.StatusOK, "<html><body><h1>Unauthorized</h1></body></html>")
			}

			// Check if the user's username is different
			username, err := repo.GetUsername(ctx, user.Identity.Id)
			if err != nil {
				return err
			}

			if user.Traits.Username != username {
				// If username is different, update it in the user meta table
				if _, err := repo.Update(ctx, user.Identity.Id, domain.UpdateArgs{
					"Username": user.Traits.Username,
				}); err != nil {
					return err
				}
			}

			c.Set("user", user)

			return next(c)
		}
	}
}

// NewActivationMiddleware returns a middleware function which enforces account activation by denying access to users
// whose accounts are deactivated. Must be placed in the chain after the top layer protect middleware since the auth user
// is pulled from the echo context.
func NewActivationMiddleware(repo domain.UserMetaRepo) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*domain.AuthUser)
			if !ok {
				return errors.New("could not cast user to *domain.AuthUser")
			}

			ctx := c.Request().Context()

			// If this user is deactivated, deny access.
			isDeactivated, err := repo.IsDeactivated(ctx, user.Identity.Id)
			if err != nil {
				return err
			}

			// Check if the user's username is different
			username, err := repo.GetUsername(ctx, user.Identity.Id)
			if err != nil {
				return err
			}

			if user.Traits.Username != username {
				// If username is different, update it in the user meta table
				if _, err := repo.Update(ctx, user.Identity.Id, domain.UpdateArgs{
					"Username": user.Traits.Username,
				}); err != nil {
					return err
				}
			}

			if isDeactivated {
				return c.JSON(http.StatusUnauthorized, unauthorized)
			}

			return next(c)
		}
	}
}

func NewAPIProtectMiddleware(config *conf.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookiePresent := false
			bearerPresent := false

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				bearerPresent = true
			}

			sessionCookie, err := c.Cookie("ory_kratos_session")
			if err == nil && sessionCookie != nil {
				cookiePresent = true
			}

			if !cookiePresent && !bearerPresent {
				return c.JSON(http.StatusUnauthorized, unauthorized)
			}

			toSessionURL := fmt.Sprintf("%s/sessions/whoami", config.KratosPublic)

			req, err := http.NewRequest("GET", toSessionURL, nil)
			if err != nil {
				return err
			}

			// Set auth proof. Cookie takes priority, bearer token is used as fallback.
			if cookiePresent {
				for _, cookie := range c.Cookies() {
					req.AddCookie(cookie)
				}
			} else if bearerPresent {
				req.Header.Set("Authorization", authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}

			if resp.StatusCode != http.StatusOK {
				return c.JSON(http.StatusUnauthorized, unauthorized)
			}

			session := &kratos.Session{}
			if err := json.NewDecoder(resp.Body).Decode(session); err != nil {
				return err
			}

			// Ensure that the user's account has at least one verifiable address verified
			var accountVerified = false
			for _, address := range session.Identity.VerifiableAddresses {
				if address.Verified {
					accountVerified = true
					break
				}
			}

			if !accountVerified {
				return c.JSON(http.StatusUnauthorized, &domain.Response{
					Success: false,
					Message: "You must verify your account before accessing Refractor",
				})
			}

			traitBytes, err := json.Marshal(session.Identity.Traits)
			if err != nil {
				return err
			}

			traits := &domain.Traits{}
			if err := json.Unmarshal(traitBytes, traits); err != nil {
				return err
			}

			user := &domain.AuthUser{
				Traits:  traits,
				Session: session,
			}

			c.Set("user", user)

			return next(c)
		}
	}
}
