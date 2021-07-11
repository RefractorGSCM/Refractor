package http

import (
	"Refractor/domain"
	"github.com/labstack/echo/v4"
)

type serverHandler struct {
	service domain.ServerService
}

func ApplyServerHandler(apiGroup *echo.Group, s domain.ServerService, authorizer domain.Authorizer, protect echo.MiddlewareFunc) {
	handler := &serverHandler{
		service: s,
	}

	// Create the server routing group
	serverGroup := apiGroup.Group("/servers")

	serverGroup.GET("/", handler.GetServers, protect)
}

// GetServers is the route handler for /api/v1/servers
// It returns a JSON array containing all servers which the requesting user has access to.
func (h *serverHandler) GetServers(c echo.Context) error {
	return c.String(200, "ok")
}
