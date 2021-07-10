package service

import "Refractor/domain"

type gameService struct {
	games map[string]domain.Game
}

func NewGameService() domain.GameService {
	return &gameService{
		games: map[string]domain.Game{},
	}
}

func (s *gameService) AddGame(game domain.Game) {
	s.games[game.GetName()] = game
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
