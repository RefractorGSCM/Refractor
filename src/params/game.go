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
	"Refractor/domain"
	"net/http"
)

type SetGameSettingsParams struct {
	InfractionCreate *domain.InfractionCommands `json:"create"`
	InfractionUpdate *domain.InfractionCommands `json:"update"`
	InfractionDelete *domain.InfractionCommands `json:"delete"`
	InfractionRepeal *domain.InfractionCommands `json:"repeal"`
}

type cmdFieldErrBody struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

func buildManualError(field, nested string, errBody interface{}) error {
	err := &domain.HTTPError{
		Success: false,
		Message: "Input errors exist",
		Status:  http.StatusBadRequest,
	}

	err.ValidationErrors = map[string]interface{}{}

	if nested != "" {
		err.ValidationErrors = map[string]map[string]interface{}{
			field: {
				nested: errBody,
			},
		}
	} else {
		err.ValidationErrors = map[string]interface{}{
			field: errBody,
		}
	}

	return err
}

func (body SetGameSettingsParams) Validate() error {
	// Validate existence of main fields
	if body.InfractionCreate == nil {
		return buildManualError("create", "", "this field is required")
	}

	if body.InfractionUpdate == nil {
		return buildManualError("update", "", "this field is required")
	}

	if body.InfractionDelete == nil {
		return buildManualError("delete", "", "this field is required")
	}

	if body.InfractionRepeal == nil {
		return buildManualError("repeal", "", "this field is required")
	}

	// Validate commands
	if err := validateActCmds(body.InfractionCreate, "create"); err != nil {
		return err
	}

	if err := validateActCmds(body.InfractionUpdate, "update"); err != nil {
		return err
	}

	if err := validateActCmds(body.InfractionDelete, "delete"); err != nil {
		return err
	}

	if err := validateActCmds(body.InfractionRepeal, "repeal"); err != nil {
		return err
	}

	return nil
}

func validateActCmds(cmds *domain.InfractionCommands, act string) error {
	if err := validateCmdArr(cmds.Warn, act, "warn"); err != nil {
		return err
	}
	if err := validateCmdArr(cmds.Mute, act, "warn"); err != nil {
		return err
	}
	if err := validateCmdArr(cmds.Kick, act, "warn"); err != nil {
		return err
	}
	if err := validateCmdArr(cmds.Ban, act, "warn"); err != nil {
		return err
	}

	return nil
}

func validateCmdArr(arr []string, act, infr string) error {
	for idx, cmd := range arr {
		if len(cmd) < 1 || len(cmd) > 256 {
			return buildManualError(act, infr, &cmdFieldErrBody{
				Index:   idx,
				Message: "length must be between 1 and 256",
			})
		}
	}

	return nil
}
