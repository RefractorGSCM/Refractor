package middleware

import (
	"Refractor/domain"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Enforcer struct {
	authorizer domain.Authorizer
	dom        string
	obj        string
}

func NewEnforcer(authorizer domain.Authorizer, domain string, object string) *Enforcer {
	return &Enforcer{
		authorizer: authorizer,
		dom:        domain,
		obj:        object,
	}
}

func (e *Enforcer) Enforce(action string) echo.MiddlewareFunc {
	return Enforce(e.authorizer, e.dom, e.obj, action)
}

// Enforce must be in the chain after AttachUserInfo as it lies on the user context field.
// Enforce checks if a user has permission to perform an action (act) on a object (obj) within a domain (dom).
// If they do not, it notifies them that they are unauthorized.
func Enforce(authorizer domain.Authorizer, dom, obj, act string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get("user").(*domain.AuthUser)

			fmt.Println(user.Identity.Id)

			hasPermission, err := authorizer.HasPermission(user.Identity.Id, dom, obj, act)
			if err != nil {
				return err
			}

			if !hasPermission {
				return c.JSON(http.StatusUnauthorized, &domain.Response{
					Success: false,
					Message: "You do not have permission to perform this action",
				})
			}

			return next(c)
		}
	}
}
