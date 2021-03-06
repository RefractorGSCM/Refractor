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
	validation "github.com/go-ozzo/ozzo-validation"
	"math"
	"net/http"
)

type SetGameCommandSettingsParams struct {
	InfractionCreate *domain.InfractionCommands `json:"create"`
	InfractionUpdate *domain.InfractionCommands `json:"update"`
	InfractionDelete *domain.InfractionCommands `json:"delete"`
	InfractionRepeal *domain.InfractionCommands `json:"repeal"`
	InfractionSync   *domain.InfractionCommands `json:"sync"`
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

func (body SetGameCommandSettingsParams) Validate() error {
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

	if body.InfractionSync == nil {
		return buildManualError("sync", "", "this field is required")
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

	// Ensure that warn and kick sync commands are nil since we don't support warn/kick syncing
	body.InfractionSync.Warn = nil
	body.InfractionSync.Kick = nil
	// Validate sync manually since it's treated specially
	if err := validateCmdArr(body.InfractionSync.Ban, "sync", "ban"); err != nil {
		return err
	}
	if err := validateCmdArr(body.InfractionSync.Mute, "sync", "mute"); err != nil {
		return err
	}

	return nil
}

func validateActCmds(cmds *domain.InfractionCommands, act string) error {
	if err := validateCmdArr(cmds.Warn, act, "warn"); err != nil {
		return err
	}
	if err := validateCmdArr(cmds.Mute, act, "mute"); err != nil {
		return err
	}
	if err := validateCmdArr(cmds.Kick, act, "kick"); err != nil {
		return err
	}
	if err := validateCmdArr(cmds.Ban, act, "ban"); err != nil {
		return err
	}

	return nil
}

func validateCmdArr(arr []*domain.InfractionCommand, act, infr string) error {
	for idx, cmd := range arr {
		if len(cmd.Command) < 1 || len(cmd.Command) > 256 {
			return buildManualError(act, infr, &cmdFieldErrBody{
				Index:   idx,
				Message: "length must be between 1 and 256",
			})
		}
	}

	return nil
}

type SetGameGeneralSettingsParams struct {
	EnableBanSync             bool `json:"enable_ban_sync"`
	EnableMuteSync            bool `json:"enable_mute_sync"`
	PlayerInfractionThreshold int  `json:"player_infraction_threshold"`
	PlayerInfractionTimespan  int  `json:"player_infraction_timespan"`
}

func (body SetGameGeneralSettingsParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.PlayerInfractionThreshold, validation.Required, validation.Min(0), validation.Max(math.MaxInt32)),
		validation.Field(&body.PlayerInfractionTimespan, validation.Required, validation.Min(0), validation.Max(math.MaxInt32)),
	)
}
