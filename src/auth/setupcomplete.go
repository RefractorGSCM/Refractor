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

package auth

import (
	"Refractor/domain"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h *publicHandlers) SetupCompleteHandler(c echo.Context) error {
	user, ok := c.Get("user").(*domain.AuthUser)
	if !ok {
		return fmt.Errorf("could not get AuthUser from context")
	}

	// Check if this user has a verified address
	verified := false
	for _, address := range user.Identity.VerifiableAddresses {
		if address.Verified {
			verified = true
			break
		}
	}

	if !verified {
		// If this user is not yet verified, redirect them to the frontend application where they will be notified that
		// they have to verify their email.
		return c.JSON(http.StatusTemporaryRedirect, h.config.FrontendRoot)
	}

	return c.Render(http.StatusOK, "setupcomplete", h.config.FrontendRoot)
}
