package http

import (
	"Refractor/domain"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type groupHandler struct {
	service domain.GroupService
}

func ApplyGroupHandler(apiGroup *echo.Group, s domain.GroupService, authorizer domain.Authorizer, protect echo.MiddlewareFunc) {
	handler := &groupHandler{
		service: s,
	}

	// Create the server routing group
	groupGroup := apiGroup.Group("/groups")

	groupGroup.GET("/", handler.GetGroups, protect)
}

func (h *groupHandler) GetGroups(c echo.Context) error {
	groups := []*domain.Group{
		{
			ID:          1,
			Name:        "Super Admin",
			Color:       0xff0000,
			Position:    1,
			Permissions: "1",
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
		},
		{
			ID:          2,
			Name:        "Admin",
			Color:       0xff4d00,
			Position:    2,
			Permissions: "2",
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
		},
		{
			ID:          3,
			Name:        "Moderator",
			Color:       0x00ff11,
			Position:    3,
			Permissions: "4",
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
		},
		{
			ID:          4,
			Name:        "Everyone",
			Color:       0xe3e3e3,
			Position:    4,
			Permissions: "1",
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
		},
	}

	return c.JSON(http.StatusOK, &domain.Response{
		Success: true,
		Message: fmt.Sprintf("Fetched %d groups", len(groups)),
		Payload: groups,
	})
}
