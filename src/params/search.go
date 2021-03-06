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
	"Refractor/params/rules"
	"Refractor/params/validators"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"math"
	"strings"
)

type SearchParams struct {
	Offset int `json:"offset" form:"offset"`
	Limit  int `json:"limit" form:"limit"`
}

func (body SearchParams) Validate() error {
	return ValidateStruct(&body,
		validation.Field(&body.Offset, rules.SearchOffsetRules...),
		validation.Field(&body.Limit, rules.SearchLimitRules.Prepend(validation.Required)...),
	)
}

type SearchPlayerParams struct {
	Term     string `json:"term" form:"term"`
	Type     string `json:"type" form:"type"`
	Platform string `json:"platform" form:"platform"`
	*SearchParams
}

var validPlayerSearchTypes = []string{"name", "id"}

func (body SearchPlayerParams) Validate() error {
	if body.SearchParams == nil {
		return fmt.Errorf("no search params provided")
	}
	if err := body.SearchParams.Validate(); err != nil {
		return err
	}

	body.Term = strings.TrimSpace(body.Term)
	body.Type = strings.TrimSpace(body.Type)

	return ValidateStruct(&body,
		validation.Field(&body.Term, validation.Required, validation.Length(1, 128)),
		validation.Field(&body.Type, validation.Required, validation.By(validators.ValueInStrArray(validPlayerSearchTypes))),
		validation.Field(&body.Platform, validation.By(validators.ValueInStrArray(domain.AllPlatforms)),
			validation.By(func(value interface{}) error {
				// if body.Type is set to "id", then platform is required
				if body.Type != "id" {
					return nil
				}

				platform, ok := value.(string)
				if !ok || (ok && len(strings.TrimSpace(platform)) == 0) {
					return errors.New("platform is required if search type is id")
				}

				return nil
			})),
	)
}

type SearchInfractionParams struct {
	Type     *string `json:"type" form:"type"`
	Game     *string `json:"game" form:"game"`
	PlayerID *string `json:"player_id" form:"player_id"`
	Platform *string `json:"platform" form:"platform"`
	ServerID *int64  `json:"server_id" form:"server_id"`
	UserID   *string `json:"user_id" form:"user_id"`
	*SearchParams
}

var validInfractionTypes = []string{domain.InfractionTypeWarning, domain.InfractionTypeMute, domain.InfractionTypeKick, domain.InfractionTypeBan}

func (body SearchInfractionParams) Validate() error {
	if body.SearchParams == nil {
		return fmt.Errorf("no search params provided")
	}
	if err := body.SearchParams.Validate(); err != nil {
		return err
	}

	return ValidateStruct(&body,
		validation.Field(&body.Type, validation.By(validators.PtrValueInStrArray(validInfractionTypes))),
		validation.Field(&body.Game, validation.By(validators.PtrValueInStrArray(domain.AllGames))),
		validation.Field(&body.PlayerID, rules.PlayerIDRules...),
		validation.Field(&body.Platform, validation.By(validators.PtrValueInStrArray(domain.AllPlatforms)),
			validation.By(func(value interface{}) error {
				// if body.PlayerID is set then platform is required
				if body.PlayerID == nil {
					return nil
				}

				platformPtr, ok := value.(*string)
				if !ok || platformPtr == nil {
					return errors.New("platform is required if player_id is set")
				}

				platform := *platformPtr
				if len(strings.TrimSpace(platform)) == 0 {
					return errors.New("platform is required if player_id is set")
				}

				return nil
			})),
		validation.Field(&body.UserID, rules.UserIDRules...),
	)
}

type SearchMessagesParams struct {
	PlayerID  *string `json:"player_id" form:"player_id"`
	Platform  *string `json:"platform" form:"platform"`
	ServerID  *int64  `json:"server_id" form:"server_id"`
	Game      *string `json:"game" form:"game"`
	StartDate *int64  `json:"start_date" form:"start_date"`
	EndDate   *int64  `json:"end_date" form:"end_date"`
	Query     *string `json:"query" form:"query"`
	*SearchParams
}

func (body SearchMessagesParams) Validate() error {
	if body.SearchParams == nil {
		return fmt.Errorf("no search params provided")
	}
	if err := body.SearchParams.Validate(); err != nil {
		return err
	}

	return ValidateStruct(&body,
		validation.Field(&body.PlayerID, rules.PlayerIDRules...),
		validation.Field(&body.Platform, validation.By(validators.PtrValueInStrArray(domain.AllPlatforms)),
			validation.By(func(value interface{}) error {
				// if body.PlayerID is set then platform is required
				if body.PlayerID == nil {
					return nil
				}

				platformPtr, ok := value.(*string)
				if !ok || platformPtr == nil {
					return errors.New("platform is required if player_id is set")
				}

				platform := *platformPtr
				if len(strings.TrimSpace(platform)) == 0 {
					return errors.New("platform is required if player_id is set")
				}

				return nil
			})),
		validation.Field(&body.ServerID, validation.Min(1), validation.Max(math.MaxInt32)),
		validation.Field(&body.Game, validation.By(validators.PtrValueInStrArray(domain.AllGames))),
		validation.Field(&body.StartDate, validation.Min(1), validation.Max(math.MaxInt64),
			validation.By(func(value interface{}) error {
				// if body.EndDate is set then StartDate is required
				if body.EndDate == nil {
					return nil
				}

				startDatePtr, ok := value.(*int64)
				if !ok || startDatePtr == nil {
					return errors.New("start_date is required if start_date is set")
				}

				return nil
			})),
		validation.Field(&body.EndDate, validation.Min(1), validation.Max(math.MaxInt64),
			validation.By(func(value interface{}) error {
				// if body.StartDate is set then EndDate is required
				if body.StartDate == nil {
					return nil
				}

				endDatePtr, ok := value.(*int64)
				if !ok || endDatePtr == nil {
					return errors.New("end_date is required if start_date is set")
				}

				return nil
			})),
		validation.Field(&body.Query, validation.Length(0, 128)),
	)
}
