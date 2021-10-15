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

package service

import (
	"Refractor/domain"
	"time"
)

type gameService struct {
	repo    domain.GameRepo
	timeout time.Duration
	games   map[string]domain.Game
}

func NewGameService(gr domain.GameRepo, to time.Duration) domain.GameService {
	return &gameService{
		repo:    gr,
		timeout: to,
		games:   map[string]domain.Game{},
	}
}

func (s *gameService) AddGame(game domain.Game) {
	s.games[game.GetName()] = game

	domain.AllGames = append(domain.AllGames, game.GetName())
}

func (s *gameService) GetAllGames() []domain.Game {
	var games []domain.Game

	for _, game := range s.games {
		games = append(games, game)
	}

	return games
}

func (s *gameService) GameExists(name string) bool {
	return s.games[name] != nil
}

func (s *gameService) GetGame(name string) (domain.Game, error) {
	if !s.GameExists(name) {
		return nil, domain.ErrNotFound
	}

	return s.games[name], nil
}

func (s *gameService) GetGameSettings(game domain.Game) (*domain.GameSettings, error) {
	settings, err := s.repo.GetSettings(game)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (s *gameService) SetGameSettings(game domain.Game, settings *domain.GameSettings) error {
	if err := s.repo.SetSettings(game, settings); err != nil {
		return err
	}

	return nil
}
