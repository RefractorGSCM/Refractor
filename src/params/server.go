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

package params

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"strings"
)

type CreateServerParams struct {
	Game         string `form:"game" json:"game"`
	Name         string `form:"name" json:"name"`
	Address      string `form:"address" json:"address"`
	RCONPort     string `form:"rcon_port" json:"rcon_port"`
	RCONPassword string `form:"rcon_password" json:"rcon_password"`
}

func (body CreateServerParams) Validate() error {
	body.Game = strings.TrimSpace(body.Game)
	body.Name = strings.TrimSpace(body.Name)
	body.Address = strings.TrimSpace(body.Address)
	body.RCONPort = strings.TrimSpace(body.RCONPort)
	body.RCONPassword = strings.TrimSpace(body.RCONPassword)

	return ValidateStruct(&body,
		validation.Field(&body.Game, validation.Required, validation.Length(1, 32)),
		validation.Field(&body.Name, validation.Required, validation.Length(1, 20)),
		validation.Field(&body.Address, validation.Required, is.IPv4),
		validation.Field(&body.RCONPort, validation.Required, is.Port),
		validation.Field(&body.RCONPassword, validation.Required, validation.Length(1, 128)),
	)
}

type UpdateServerParams struct {
	Game         *string `json:"game" form:"game"`
	Name         *string `json:"name" form:"name"`
	Address      *string `json:"address" form:"address"`
	RCONPort     *string `json:"rcon_port" form:"rcon_port"`
	RCONPassword *string `json:"rcon_password" form:"rcon_password"`
}

func (body UpdateServerParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.Game, validation.By(stringPointerNotEmpty), validation.Length(1, 32)),
		validation.Field(&body.Name, validation.By(stringPointerNotEmpty), validation.Length(1, 20)),
		validation.Field(&body.Address, validation.By(stringPointerNotEmpty), is.IPv4),
		validation.Field(&body.RCONPort, validation.By(stringPointerNotEmpty), is.Port),
		validation.Field(&body.RCONPassword, validation.By(stringPointerNotEmpty), validation.Length(1, 128)),
	)
}
