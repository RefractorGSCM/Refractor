package auth

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h *publicHandlers) RootHandler(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect,
		fmt.Sprintf("%s/self-service/login/browser", h.config.KratosPublic))
}
