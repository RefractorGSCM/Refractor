package middleware

import (
	"Refractor/domain"
	"Refractor/pkg/conf"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	kratos "github.com/ory/kratos-client-go"
	"net/http"
)

func NewProtectMiddleware(config *conf.Config) echo.MiddlewareFunc {
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

			c.Set("user", &domain.AuthUser{
				Traits:  traits,
				Session: session,
			})

			return next(c)
		}
	}
}
