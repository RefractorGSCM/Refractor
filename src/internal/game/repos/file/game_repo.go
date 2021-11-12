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
	"encoding/json"
	"fmt"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"time"
)

const opTag = "GameRepo.File."

type gameRepo struct {
	cache  *gocache.Cache
	logger *zap.Logger
}

func NewGameRepo(log *zap.Logger) domain.GameRepo {
	return &gameRepo{
		cache:  gocache.New(time.Hour, time.Hour),
		logger: log,
	}
}

func (r *gameRepo) GetSettings(game domain.Game) (*domain.GameSettings, error) {
	const op = opTag + "GetSettings"

	// Check if this game's settings exists in the cache. If they do, return them and skip the IO.
	if st, found := r.cache.Get(game.GetName()); found {
		settings := st.(*domain.GameSettings)

		return settings, nil
	}

	// Check if data file exists
	if _, err := os.Stat(fmt.Sprintf("./data/%s_settings.json", game.GetName())); os.IsNotExist(err) {
		r.logger.Info("Game settings file does not exist. Creating from defaults...", zap.String("Game", game.GetName()))
		// If it doesn't, use SetSettings to create it
		if err := r.SetSettings(game, game.GetDefaultSettings()); err != nil {
			r.logger.Error("Could not create game settings file", zap.String("Game", game.GetName()), zap.Error(err))
			return nil, errors.Wrap(err, op)
		}
	}

	// Open data file and decode the data within
	file, err := os.Open(fmt.Sprintf("./data/%s_settings.json", game.GetName()))
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	defer func() {
		_ = file.Close()
	}()

	decoder := json.NewDecoder(file)

	decodedSettings := &domain.GameSettings{}
	if err := decoder.Decode(decodedSettings); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Set game settings in cache
	r.cache.SetDefault(game.GetName(), decodedSettings)

	return decodedSettings, nil
}

func (r *gameRepo) SetSettings(game domain.Game, settings *domain.GameSettings) error {
	const op = opTag + "SetSettings"

	// Check if data directory exists
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		if err := os.Mkdir("./data", os.ModePerm); err != nil {
			return errors.Wrap(err, op)
		}
	}

	// Create data file
	file, err := os.Create(fmt.Sprintf("./data/%s_settings.json", game.GetName()))
	if err != nil {
		return errors.Wrap(err, op)
	}

	defer func() {
		_ = file.Close()
	}()

	// JSON encode the settings struct
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return errors.Wrap(err, op)
	}

	// Write to the file
	if _, err := file.Write(data); err != nil {
		return errors.Wrap(err, op)
	}

	// Update cache
	r.cache.SetDefault(game.GetName(), settings)

	return nil
}
