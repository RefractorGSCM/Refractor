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

package file

import (
	"Refractor/domain"
	"context"
	"encoding/gob"
	"fmt"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"os"
	"time"
)

const opTag = "GameRepo.File."

type gameRepo struct {
	cache *gocache.Cache
}

func NewGameRepo() domain.GameRepo {
	return &gameRepo{
		cache: gocache.New(time.Hour, time.Hour),
	}
}

func (r *gameRepo) GetSettings(ctx context.Context, game domain.Game) (*domain.GameSettings, error) {
	const op = opTag + "GetSettings"

	// Check if this game's settings exists in the cache. If they do, return them and skip the IO.
	if st, found := r.cache.Get(game.GetName()); found {
		settings := st.(*domain.GameSettings)

		return settings, nil
	}

	// Check if data file exists
	if _, err := os.Stat(fmt.Sprintf("./data/%s_settings.gob", game.GetName())); os.IsNotExist(err) {
		// If it doesn't, use SetSettings to create it
		if err := r.SetSettings(ctx, game, game.GetDefaultSettings()); err != nil {
			return nil, errors.Wrap(err, op)
		}
	}

	// Open data file and decode the data within
	file, err := os.Open(fmt.Sprintf("./data/%s_settings.gob", game.GetName()))
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	defer func() {
		_ = file.Close()
	}()

	decoder := gob.NewDecoder(file)

	decodedSettings := &domain.GameSettings{}
	if err := decoder.Decode(decodedSettings); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Set game settings in cache
	r.cache.SetDefault(game.GetName(), decodedSettings)

	return decodedSettings, nil
}

func (r *gameRepo) SetSettings(ctx context.Context, game domain.Game, settings *domain.GameSettings) error {
	const op = opTag + "SetSettings"

	// Check if data directory exists
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		if err := os.Mkdir("./data", os.ModePerm); err != nil {
			return errors.Wrap(err, op)
		}
	}

	// Create data file
	file, err := os.Create(fmt.Sprintf("./data/%s_settings.gob", game.GetName()))
	if err != nil {
		return errors.Wrap(err, op)
	}

	defer func() {
		_ = file.Close()
	}()

	// Gob encode the settings struct
	encoder := gob.NewEncoder(file)

	if err := encoder.Encode(settings); err != nil {
		return errors.Wrap(err, op)
	}

	// Update cache
	r.cache.SetDefault(game.GetName(), settings)

	return nil
}
